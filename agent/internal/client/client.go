// Package client handles backend communication
package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"smart-agent/internal/collector"
	"smart-agent/internal/config"
	"smart-agent/internal/identity"
	pb "smart-monitor/pbtypes/monitor"
)

// Client handles communication with backend
type Client struct {
	config      *config.Config
	identity    *identity.Manager
	collector   *collector.Collector
	conn        *grpc.ClientConn
	grpcClient  pb.MonitorServiceClient
	credentials *identity.Credentials
}

// NewClient creates a new backend client
func NewClient(cfg *config.Config, identityMgr *identity.Manager, metricsCollector *collector.Collector) *Client {
	return &Client{
		config:    cfg,
		identity:  identityMgr,
		collector: metricsCollector,
	}
}

// Connect establishes connection to backend
func (c *Client) Connect(ctx context.Context) error {
	log.Printf("Connecting to backend at %s...", c.config.BackendAddr)

	// TODO: Add TLS support when BackendTLS is true
	conn, err := grpc.DialContext(
		ctx,
		c.config.BackendAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.grpcClient = pb.NewMonitorServiceClient(conn)
	log.Printf("✓ Connected to backend")

	return nil
}

// Register registers the agent with backend
func (c *Client) Register(ctx context.Context) error {
	log.Println("Registering agent with backend...")

	req := &pb.RegisterRequest{
		Hostname:     c.config.Hostname,
		IpAddress:    c.config.IPAddress,
		AgentVersion: c.config.AgentVersion,
		Metadata:     c.config.Metadata,
	}

	resp, err := c.grpcClient.RegisterAgent(ctx, req)
	if err != nil {
		return fmt.Errorf("registration RPC failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	// Store credentials
	c.credentials = &identity.Credentials{
		AgentID:     resp.AgentId,
		AccessToken: resp.AccessToken,
		ExpiresAt:   resp.ExpiresAt,
		Hostname:    c.config.Hostname,
		IPAddress:   c.config.IPAddress,
	}

	// Save to disk
	if err := c.identity.SaveCredentials(c.credentials); err != nil {
		log.Printf("Warning: Failed to save credentials: %v", err)
	}

	log.Printf("✓ Agent registered successfully")
	log.Printf("  Agent ID: %s", c.credentials.AgentID)
	log.Printf("  Token expires: %s", time.Unix(c.credentials.ExpiresAt, 0).Format(time.RFC3339))

	return nil
}

// LoadOrRegister loads existing credentials or registers new agent
func (c *Client) LoadOrRegister(ctx context.Context) error {
	// Try to load existing credentials
	if c.identity.HasValidCredentials() {
		creds, err := c.identity.LoadCredentials()
		if err == nil {
			c.credentials = creds
			log.Printf("✓ Loaded existing credentials for agent %s", creds.AgentID)
			return nil
		}
	}

	// Register new agent
	return c.Register(ctx)
}

// StreamMetrics streams metrics to backend
func (c *Client) StreamMetrics(ctx context.Context) error {
	if c.credentials == nil {
		return fmt.Errorf("not registered, credentials missing")
	}

	log.Println("Starting metrics streaming...")

	stream, err := c.grpcClient.StreamStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	ticker := time.NewTicker(c.config.MetricsInterval)
	defer ticker.Stop()

	log.Printf("📊 Monitoring system metrics (interval: %v)...", c.config.MetricsInterval)

	for {
		select {
		case <-ctx.Done():
			// Graceful shutdown
			resp, err := stream.CloseAndRecv()
			if err != nil {
				log.Printf("Error closing stream: %v", err)
			} else {
				log.Printf("Final response: %s", resp.Message)
			}
			return ctx.Err()

		case <-ticker.C:
			if err := c.sendMetrics(stream); err != nil {
				log.Printf("Error sending metrics: %v", err)
				return err
			}
		}
	}
}

// sendMetrics collects and sends metrics
func (c *Client) sendMetrics(stream pb.MonitorService_StreamStatsClient) error {
	// Collect metrics
	metrics, err := c.collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Build request
	req := &pb.StatsRequest{
		Hostname:     c.config.Hostname,
		AgentId:      c.credentials.AgentID,
		IpAddress:    c.config.IPAddress,
		AgentVersion: c.config.AgentVersion,
		AccessToken:  c.credentials.AccessToken,
		Cpu:          metrics.CPUPercent,
		Ram:          metrics.RAMPercent,
		Disk:         metrics.DiskPercent,
		Metadata:     c.config.Metadata,
	}

	// Send to backend
	if err := stream.Send(req); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	log.Printf("✓ Sent [%s]: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%%",
		c.credentials.AgentID, metrics.CPUPercent, metrics.RAMPercent, metrics.DiskPercent)

	return nil
}

// Close closes the connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetCredentials returns current credentials
func (c *Client) GetCredentials() *identity.Credentials {
	return c.credentials
}
