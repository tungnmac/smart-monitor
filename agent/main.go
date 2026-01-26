package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
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

	// Generate unique agent ID based on hostname and IP
	agentID := generateAgentID(hostname, ipAddress)
	log.Printf("Agent ID: %s", agentID)
	log.Printf("Agent Version: %s", agentVersion)

	// Connect to backend via gRPC
	conn, err := grpc.Dial(backendAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to backend: %v", err)
	}
	defer conn.Close()
	log.Printf("✓ Connected to backend at %s", backendAddr)

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
	if err := streamStats(ctx, client, hostname, credentials, metadata); err != nil {
		log.Printf("Error streaming stats: %v", err)
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

// streamStats collects and streams system stats to backend
func streamStats(ctx context.Context, client pb.MonitorServiceClient, hostname string, credentials *AgentCredentials, metadata map[string]string) error {
	stream, err := client.StreamStats(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("📊 Monitoring system metrics...")

	// Get IP address
	ipAddress := getLocalIP()

	for {
		select {
		case <-ctx.Done():
			// Close stream gracefully
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Printf("Error closing stream: %v", err)
			} else {
				log.Printf("Final response: %s", resp.Message)
			}
			return nil

		case <-ticker.C:
			// Collect metrics
			stats, err := collectStats(hostname, credentials.AgentID, ipAddress, credentials.AccessToken, metadata)
			if err != nil {
				log.Printf("Error collecting stats: %v", err)
				continue
			}

			// Send to backend
			if err := stream.Send(stats); err != nil {
				log.Printf("Error sending stats: %v", err)
				return err
			}

			log.Printf("✓ Sent [%s]: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%%",
				credentials.AgentID, stats.Cpu, stats.Ram, stats.Disk)
		}
	}
}

// collectStats gathers current system statistics
func collectStats(hostname, agentID, ipAddress, accessToken string, metadata map[string]string) (*pb.StatsRequest, error) {
	// CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	cpuUsage := 0.0
	if len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	// Memory usage
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Disk usage (root partition)
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
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
	}, nil
}
