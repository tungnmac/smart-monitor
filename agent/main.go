package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "smart-monitor/pbtypes/monitor"
)

const (
	backendAddr  = "localhost:50051" // gRPC backend address
	interval     = 2 * time.Second   // Monitoring interval
	agentVersion = "1.0.0"           // Agent version
	tokenFile    = ".agent_token"    // File to store access token
)

type AgentCredentials struct {
	AgentID     string
	AccessToken string
	ExpiresAt   int64
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("=== Smart Monitor Agent ===")

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	log.Printf("Agent hostname: %s", hostname)

	// Get local IP address
	ipAddress := getLocalIP()
	log.Printf("Agent IP: %s", ipAddress)
	log.Printf("Agent Version: %s", agentVersion)

	// Connect to backend via gRPC with retry
	var conn *grpc.ClientConn
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		conn, err = grpc.Dial(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			// Check if connection is actually possible
			client := pb.NewMonitorServiceClient(conn)
			_, err = client.RegisterAgent(ctx, &pb.RegisterRequest{Hostname: "ping"})
			if err == nil || strings.Contains(err.Error(), "already exists") {
				log.Printf("✓ Connected to backend at %s", backendAddr)
				break
			}
		}
		log.Printf("Waiting for backend at %s (attempt %d/10)...", backendAddr, i+1)
		time.Sleep(2 * time.Second)
		if i == 9 {
			log.Fatalf("Failed to connect to backend after 10 attempts: %v", err)
		}
	}
	defer conn.Close()

	client := pb.NewMonitorServiceClient(conn)

	// Metadata for agent
	metadata := map[string]string{
		"location":    "datacenter-01",
		"environment": "production",
		"os":          "linux",
	}

	// Setup context for operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register agent with backend
	log.Println("📝 Registering agent with backend...")
	credentials, err := registerAgent(ctx, client, hostname, ipAddress, metadata)
	if err != nil {
		log.Fatalf("Failed to register agent: %v", err)
	}
	log.Printf("✓ Agent registered successfully")
	log.Printf("  Agent ID: %s", credentials.AgentID)
	log.Printf("  Token expires: %s", time.Unix(credentials.ExpiresAt, 0).Format(time.RFC3339))

	// Save credentials to file
	if err := saveCredentials(credentials); err != nil {
		log.Printf("Warning: Failed to save credentials: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n⚠ Shutting down agent...")
		cancel()
	}()

	// Start monitoring and streaming stats
	log.Println("✓ Starting monitoring...")
	for {
		err := streamMetrics(ctx, client, hostname, ipAddress, credentials, metadata)
		if err != nil {
			log.Printf("Error streaming stats: %v", err)

			// If error is unauthenticated, wait and try to re-register
			if strings.Contains(err.Error(), "invalid token") || strings.Contains(err.Error(), "unauthenticated") {
				log.Println("⚠ Authentication failed. Re-registering in 10s...")
				time.Sleep(10 * time.Second)

				newCreds, regErr := registerAgent(ctx, client, hostname, ipAddress, metadata)
				if regErr == nil {
					credentials = newCreds
					saveCredentials(credentials)
					log.Printf("✓ Re-registered successfully. New ID: %s", credentials.AgentID)
					continue
				}
				log.Printf("Failed to re-register: %v", regErr)
			}
		}

		// Wait before retry for other errors
		time.Sleep(5 * time.Second)
		if ctx.Err() != nil {
			break
		}
	}

	log.Println("Agent stopped")
}

// generateAgentID creates a unique identifier for this agent
func generateAgentID(hostname, ipAddress string) string {
	data := fmt.Sprintf("%s-%s-%d", hostname, ipAddress, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("agent-%x", hash[:8])
}

// getLocalIP returns the local IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknown"
}

// registerAgent registers the agent with the backend
func registerAgent(ctx context.Context, client pb.MonitorServiceClient, hostname, ipAddress string, metadata map[string]string) (*AgentCredentials, error) {
	req := &pb.RegisterRequest{
		Hostname:     hostname,
		IpAddress:    ipAddress,
		AgentVersion: agentVersion,
		Metadata:     metadata,
	}

	resp, err := client.RegisterAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("registration RPC failed: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("registration failed: %s", resp.Message)
	}

	return &AgentCredentials{
		AgentID:     resp.AgentId,
		AccessToken: resp.AccessToken,
		ExpiresAt:   resp.ExpiresAt,
	}, nil
}

// saveCredentials saves agent credentials to file
func saveCredentials(creds *AgentCredentials) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFile, data, 0600)
}

// loadCredentials loads agent credentials from file
func loadCredentials() (*AgentCredentials, error) {
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}

	var creds AgentCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}

	return &creds, nil
}

// streamMetrics collects and streams system metrics to backend
func streamMetrics(ctx context.Context, client pb.MonitorServiceClient, hostname, ipAddress string, credentials *AgentCredentials, metadata map[string]string) error {
	stream, err := client.StreamStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Channel to trigger immediate heartbeat
	triggerChan := make(chan struct{}, 1)

	log.Printf("📊 Monitoring system metrics (interval: %v)...", interval)

	// Goroutine to receive commands from backend
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				if err != io.EOF {
					log.Printf("Error receiving from stream: %v", err)
				}
				return
			}

			if resp.Command != nil && resp.Command.Type != pb.Command_NONE {
				if resp.Command.Type == pb.Command_UPDATE_CONFIG {
					log.Printf("Received REFRESH command, triggering heartbeat...")
					select {
					case triggerChan <- struct{}{}:
					default:
						// Already triggered
					}
				} else {
					handleCommand(resp.Command)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Close stream gracefully
			err := stream.CloseSend()
			if err != nil {
				log.Printf("Error closing stream: %v", err)
			}
			return nil

		case <-triggerChan:
			log.Printf("Executing immediate heartbeat...")
			collectAndSend(hostname, ipAddress, credentials, metadata, stream)

		case <-ticker.C:
			collectAndSend(hostname, ipAddress, credentials, metadata, stream)
		}
	}
}

// collectAndSend collects stats and sends them through the stream
func collectAndSend(hostname, ipAddress string, credentials *AgentCredentials, metadata map[string]string, stream pb.MonitorService_StreamStatsClient) {
	// Collect metrics
	stats, err := collectStats(hostname, credentials.AgentID, ipAddress, credentials.AccessToken, metadata)
	if err != nil {
		log.Printf("Error collecting stats: %v", err)
		return
	}

	// Send to backend
	if err := stream.Send(stats); err != nil {
		log.Printf("Error sending stats: %v", err)
		return
	}

	log.Printf("✓ Sent [%s]: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%%, Procs=%d",
		credentials.AgentID, stats.Cpu, stats.Ram, stats.Disk, len(stats.Processes))
}

// handleCommand executes commands received from backend
func handleCommand(cmd *pb.Command) {
	log.Printf("Received command: %v with args: %v", cmd.Type, cmd.Args)

	switch cmd.Type {
	case pb.Command_KILL_PROCESS:
		pidStr, ok := cmd.Args["pid"]
		if !ok {
			log.Printf("Error: PID argument missing for KILL_PROCESS command")
			return
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Printf("Error: Invalid PID %s: %v", pidStr, err)
			return
		}

		err = killProcess(int32(pid))
		if err != nil {
			log.Printf("Error: Failed to kill process %d: %v", pid, err)
		} else {
			log.Printf("✓ Process %d killed successfully", pid)
		}

	case pb.Command_RESTART_AGENT:
		log.Printf("Restarting agent...")
		// In a real implementation, this might exit and let a service manager restart it
		os.Exit(0)

	default:
		log.Printf("Unhandled command type: %v", cmd.Type)
	}
}

// killProcess kills a process by PID
func killProcess(pid int32) error {
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}

// collectStats gathers current system statistics
func collectStats(hostname, agentID, ipAddress, accessToken string, metadata map[string]string) (*pb.StatsRequest, error) {
	// CPU usage
	cpuUsage := 0.0
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		// Only log this warning once or less frequently? For now, let's keep it but make it clear.

		// Fallback for macOS (Darwin) when CGO is disabled
		if runtime.GOOS == "darwin" {
			// Using 'top' to get CPU usage: "CPU usage: 5.34% user, 4.41% sys, 90.24% idle"
			// Using a more robust command to get just the numbers
			out, execErr := exec.Command("sh", "-c", "top -l 1 -n 0 | grep 'CPU usage'").Output()
			if execErr == nil {
				str := string(out)
				// Format: "CPU usage: 5.34% user, 4.41% sys, 90.24% idle"
				userParts := strings.Split(str, "user")
				if len(userParts) > 0 {
					userStr := userParts[0]
					if idx := strings.LastIndex(userStr, ":"); idx != -1 {
						userStr = userStr[idx+1:]
					}
					userStr = strings.TrimSpace(strings.ReplaceAll(userStr, "%", ""))
					userVal, _ := strconv.ParseFloat(userStr, 64)

					sysParts := strings.Split(str, "sys")
					if len(sysParts) > 0 {
						sysStr := sysParts[0]
						if idx := strings.LastIndex(sysStr, ","); idx != -1 {
							sysStr = sysStr[idx+1:]
						}
						sysStr = strings.TrimSpace(strings.ReplaceAll(sysStr, "%", ""))
						sysVal, _ := strconv.ParseFloat(sysStr, 64)

						cpuUsage = userVal + sysVal
					}
				}
			}
		}
	} else if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	} // Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Disk usage (root partition)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	// Collect top processes
	var pbProcesses []*pb.ProcessInfo
	procs, err := process.Processes()
	if err == nil {
		type procWithStats struct {
			p    *process.Process
			cpu  float64
			mem  float32
			name string
		}
		var stats []procWithStats
		for _, p := range procs {
			cpuP, _ := p.CPUPercent()
			memP, _ := p.MemoryPercent()
			name, _ := p.Name()
			if name == "" {
				name = fmt.Sprintf("system-proc-%d", p.Pid)
			}
			stats = append(stats, procWithStats{p, cpuP, memP, name})
		}

		// Sort by CPU usage, then Memory
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].cpu != stats[j].cpu {
				return stats[i].cpu > stats[j].cpu
			}
			return stats[i].mem > stats[j].mem
		})

		limit := 200 // Increased limit to include more processes from VMs/Containers
		if len(stats) < limit {
			limit = len(stats)
		}

		for i := 0; i < limit; i++ {
			p := stats[i].p
			cmd, _ := p.Cmdline()

			// Better detection for containerized/virtualized processes
			displayName := stats[i].name
			cmdLow := strings.ToLower(cmd)
			if runtime.GOOS == "linux" {
				if cgroup, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", p.Pid)); err == nil {
					cgroupStr := string(cgroup)
					if strings.Contains(cgroupStr, "docker") || strings.Contains(cgroupStr, "kubepods") || strings.Contains(cgroupStr, "containerd") {
						displayName = "[Container] " + displayName
					}
				}
			}

			// VM detection based on common process names/commands
			if strings.Contains(cmdLow, "qemu") || strings.Contains(cmdLow, "vbox") || strings.Contains(cmdLow, "virtualbox") || strings.Contains(cmdLow, "vmware") {
				if !strings.HasPrefix(displayName, "[") {
					displayName = "[VM Host] " + displayName
				}
			}

			// Include thread information if significant
			if numThreads, err := p.NumThreads(); err == nil && numThreads > 1 {
				displayName = fmt.Sprintf("%s (%d thds)", displayName, numThreads)
			}

			var port int32
			conns, _ := p.Connections()
			for _, conn := range conns {
				if conn.Status == "LISTEN" {
					port = int32(conn.Laddr.Port)
					break
				}
			}

			pbProcesses = append(pbProcesses, &pb.ProcessInfo{
				Pid:     p.Pid,
				Name:    displayName,
				Cpu:     stats[i].cpu,
				Memory:  float64(stats[i].mem),
				Command: cmd,
				Port:    port,
			})
		}
	}

	return &pb.StatsRequest{
		Hostname:     hostname,
		AgentId:      agentID,
		IpAddress:    ipAddress,
		AgentVersion: agentVersion,
		AccessToken:  accessToken,
		Cpu:          cpuUsage,
		Ram:          memInfo.UsedPercent,
		Disk:         diskInfo.UsedPercent,
		Metadata:     metadata,
		Processes:    pbProcesses,
	}, nil
}
