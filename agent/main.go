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

	"smart-agent/internal/buffer"
	"smart-agent/internal/client"
	pb "smart-monitor/pbtypes/monitor"
)

const (
	backendAddr   = "localhost:50051" // gRPC backend address
	interval      = 2 * time.Second   // Monitoring interval
	agentVersion  = "1.0.0"           // Agent version
	tokenFile     = ".agent_token"    // File to store access token
	bufferSize    = 1000              // Ring buffer capacity
	bufferDataDir = ".agent_buffer"   // Buffer data storage directory
	batchSize     = 50                // Batch size for flushing
)

type AgentCredentials struct {
	AgentID     string
	AccessToken string
	ExpiresAt   int64
	IsEnrolled  bool   // Whether agent is enrolled and authorized to collect data
	PolicyID    string // Assigned policy ID
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

	// Initialize ring buffer
	log.Println("📦 Initializing ring buffer...")
	rb, err := buffer.NewRingBuffer(bufferSize, bufferDataDir)
	if err != nil {
		log.Fatalf("Failed to initialize ring buffer: %v", err)
	}
	log.Printf("✓ Ring buffer initialized (capacity: %d)", bufferSize)

	// Start persistence worker
	rb.StartPersistenceWorker(5 * time.Second)
	log.Println("✓ Buffer persistence worker started")

	// Connect to backend via gRPC with retry and circuit breaker
	log.Println("🌐 Initializing connection manager with retry and circuit breaker...")
	connMgr := client.NewConnectionManager(backendAddr, client.DefaultRetryConfig())
	conn, err := connMgr.Connect()
	if err != nil {
		log.Printf("⚠️  Failed to connect to backend: %v", err)
		log.Println("⚠️  Agent will run independently and buffer metrics locally")
		log.Println("🔄 Starting background reconnection attempts...")
	} else {
		log.Printf("✓ Connected to backend at %s", backendAddr)
	}

	defer func() {
		log.Println("💾 Persisting buffer on shutdown...")
		if err := rb.ShutdownPersist(); err != nil {
			log.Printf("Error persisting buffer: %v", err)
		}
		if conn != nil {
			conn.Close()
		}
	}()

	var monitorClient pb.MonitorServiceClient
	if conn != nil {
		monitorClient = pb.NewMonitorServiceClient(conn)
	}

	// Metadata for agent
	metadata := map[string]string{
		"location":    "datacenter-01",
		"environment": "production",
		"os":          "linux",
	}

	// Setup context for operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Step 1: Register agent with backend
	var credentials *AgentCredentials
	if monitorClient != nil {
		log.Println("📝 Step 1: Registering agent with backend center...")
		var err error
		credentials, err = registerAgent(ctx, monitorClient, hostname, ipAddress, metadata)
		if err != nil {
			log.Printf("⚠️  Failed to register agent: %v", err)
			log.Println("⚠️  Will generate local credentials and continue in standalone mode")
		} else {
			log.Printf("✅ Agent registered successfully")
			log.Printf("  Agent ID: %s", credentials.AgentID)
			log.Printf("  Token expires: %s", time.Unix(credentials.ExpiresAt, 0).Format(time.RFC3339))

			// Step 2: Enroll agent - Wait for policy assignment
			log.Println("📋 Step 2: Enrolling agent - waiting for policy authorization...")
			enrolled, policyID := enrollAgent(ctx, monitorClient, credentials.AgentID, credentials.AccessToken)
			if enrolled {
				credentials.IsEnrolled = true
				credentials.PolicyID = policyID
				log.Printf("✅ Agent enrolled successfully")
				log.Printf("  Policy ID: %s", policyID)
				log.Println("  Agent is authorized to collect and send metrics")
			} else {
				log.Println("⚠️  Agent NOT enrolled - no policy assigned")
				log.Println("  Agent will buffer metrics locally until enrolled")
				credentials.IsEnrolled = false
			}
		}
	}

	// If no credentials, generate local ones
	if credentials == nil {
		credentials = &AgentCredentials{
			AgentID:     generateAgentID(hostname, getLocalIP()),
			AccessToken: "local-token-" + fmt.Sprintf("%d", time.Now().Unix()),
			ExpiresAt:   time.Now().Add(24 * time.Hour).Unix(),
			IsEnrolled:  false,
			PolicyID:    "",
		}
		log.Printf("📝 Generated local credentials for standalone mode")
		log.Printf("  Agent ID: %s", credentials.AgentID)
	}

	// Save credentials to file
	if err := saveCredentials(credentials); err != nil {
		log.Printf("Warning: Failed to save credentials: %v", err)
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\n⚠️  Shutting down agent...")
		cancel()
	}()

	// Start background reconnection goroutine
	go startBackgroundReconnection(ctx, connMgr)

	// Start monitoring and streaming stats
	log.Println("✓ Starting monitoring...")
	if err := streamStats(ctx, monitorClient, hostname, credentials, metadata, rb, connMgr); err != nil {
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
		IsEnrolled:  false, // Not enrolled yet
		PolicyID:    "",
	}, nil
}

// enrollAgent checks if agent is enrolled and has policy assigned
// Returns (isEnrolled, policyID)
func enrollAgent(ctx context.Context, client pb.MonitorServiceClient, agentID, accessToken string) (bool, string) {
	// Poll backend to check if policy has been assigned
	maxAttempts := 10
	pollInterval := 3 * time.Second

	log.Printf("  Polling backend for policy assignment (max %d attempts, %v interval)...", maxAttempts, pollInterval)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// List policies to check if any policy is applied to this agent
		listReq := &pb.ListPoliciesRequest{}
		listResp, err := client.ListPolicies(ctx, listReq)

		if err != nil {
			log.Printf("  Attempt %d/%d: Error checking policies: %v", attempt, maxAttempts, err)
			time.Sleep(pollInterval)
			continue
		}

		// Check if any policy is assigned to this agent
		for _, policy := range listResp.Policies {
			// Check if this agent is in the policy's agent list
			for _, assignedAgentID := range policy.AgentIds {
				if assignedAgentID == agentID {
					log.Printf("  ✅ Found policy assignment: %s (%s)", policy.PolicyId, policy.Name)
					return true, policy.PolicyId
				}
			}
		}

		log.Printf("  Attempt %d/%d: No policy assigned yet, waiting...", attempt, maxAttempts)
		time.Sleep(pollInterval)
	}

	log.Printf("  ⏱️  Enrollment timeout: No policy assigned after %d attempts", maxAttempts)
	log.Println("  💡 Admin needs to apply policy from backend center")
	return false, ""
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

// startBackgroundReconnection attempts to reconnect to backend periodically
func startBackgroundReconnection(ctx context.Context, connMgr *client.ConnectionManager) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping background reconnection...")
			return
		case <-ticker.C:
			state := connMgr.GetCircuitBreakerState()
			if state != client.StateClosed {
				log.Printf("🔄 Attempting background reconnection (Circuit state: %s)...", state)
				if conn, err := connMgr.ReconnectWithBackoff(); err == nil {
					log.Printf("✅ Successfully reconnected in background!")
					conn.Close()
				}
			}
		}
	}
}

// streamStats collects and streams system stats to backend
func streamStats(ctx context.Context, monitorClient pb.MonitorServiceClient, hostname string, credentials *AgentCredentials, metadata map[string]string, rb *buffer.RingBuffer, connMgr *client.ConnectionManager) error {
	// If no client connection, just collect and buffer metrics locally
	if monitorClient == nil {
		log.Println("📊 Running in standalone mode - collecting metrics locally...")
		return collectAndBufferLocally(ctx, hostname, credentials, metadata, rb)
	}

	// Check enrollment status
	if !credentials.IsEnrolled {
		log.Println("⚠️  Agent NOT enrolled - no authorization to send metrics")
		log.Println("📦 Buffering metrics locally until policy is assigned...")
		log.Println("💡 Admin must apply policy from backend center to authorize data collection")
		return collectAndBufferLocally(ctx, hostname, credentials, metadata, rb)
	}

	log.Printf("✅ Agent is enrolled with policy: %s", credentials.PolicyID)
	log.Println("📡 Starting metrics streaming to backend...")

	stream, err := monitorClient.StreamStats(ctx)
	if err != nil {
		log.Printf("Error creating stream: %v", err)
		log.Println("⚠️  Switching to local buffering mode...")
		// Continue in standalone mode
		return collectAndBufferLocally(ctx, hostname, credentials, metadata, rb)
	}

	// Create buffer manager
	bufferMgr := buffer.NewBufferManager(rb, monitorClient, 5, 5*time.Second, batchSize)
	log.Println("✓ Buffer manager initialized")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	log.Println("📊 Monitoring system metrics...")

	// Get IP address
	ipAddress := getLocalIP()

	for {
		select {
		case <-ctx.Done():
			// Close stream gracefully
			log.Println("📤 Flushing remaining buffered data...")
			if err := bufferMgr.FlushBuffer(ctx, stream); err != nil {
				log.Printf("Warning: Failed to flush remaining buffer: %v", err)
			}

			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Printf("Error closing stream: %v", err)
			} else {
				log.Printf("Final response: %s", resp.Message)
			}
			return nil

		case <-statusTicker.C:
			// Print buffer status
			status := bufferMgr.GetBufferStatus()
			log.Printf("📊 Buffer Status: %d/%d (%.1f%%) | Connected: %v",
				status["count"], status["capacity"], status["percentage"], status["is_connected"])

		case <-ticker.C:
			// Collect metrics
			stats, err := collectStats(hostname, credentials.AgentID, ipAddress, credentials.AccessToken, metadata)
			if err != nil {
				log.Printf("Error collecting stats: %v", err)
				continue
			}

			// Send to backend (or buffer if failed)
			if err := bufferMgr.Send(ctx, stream, stats); err != nil {
				log.Printf("Error sending stats: %v", err)
				continue
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

// collectAndBufferLocally collects metrics and stores them in the ring buffer when backend is unavailable
func collectAndBufferLocally(ctx context.Context, hostname string, credentials *AgentCredentials, metadata map[string]string, rb *buffer.RingBuffer) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	ipAddress := getLocalIP()

	enrollmentStatus := "Standalone Mode"
	if !credentials.IsEnrolled && credentials.AccessToken != "" && !contains(credentials.AccessToken, "local-token") {
		enrollmentStatus = "Waiting for Policy Assignment"
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("📤 Stopping local metric collection...")
			return nil

		case <-statusTicker.C:
			// Print buffer status
			count := rb.Count()
			size := rb.GetSize()
			percentage := float64(count) / float64(size) * 100
			log.Printf("📊 Buffer Status: %d/%d (%.1f%% used) - %s", count, size, percentage, enrollmentStatus)

		case <-ticker.C:
			// Collect metrics
			stats, err := collectStats(hostname, credentials.AgentID, ipAddress, credentials.AccessToken, metadata)
			if err != nil {
				log.Printf("Error collecting stats: %v", err)
				continue
			}

			// Buffer locally
			rb.Push(stats)
			log.Printf("📦 Buffered [%s]: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%% | %s",
				credentials.AgentID, stats.Cpu, stats.Ram, stats.Disk, enrollmentStatus)
		}
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
