// Package agent implements the main agent logic
package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smart-agent/internal/client"
	"smart-agent/internal/collector"
	"smart-agent/internal/config"
	"smart-agent/internal/identity"
)

// Agent represents the monitoring agent
type Agent struct {
	config    *config.Config
	identity  *identity.Manager
	collector *collector.Collector
	client    *client.Client
	ctx       context.Context
	cancel    context.CancelFunc
}

// New creates a new agent instance
func New(cfg *config.Config) (*Agent, error) {
	// Setup identity manager
	identityMgr := identity.NewManager(cfg.TokenFile)

	// Get IP address if not configured
	if cfg.IPAddress == "" {
		cfg.IPAddress = identityMgr.GetLocalIP()
	}

	// Setup metrics collector
	metricsCollector := collector.NewCollector("/")

	// Setup backend client
	backendClient := client.NewClient(cfg, identityMgr, metricsCollector)

	// Create agent
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		config:    cfg,
		identity:  identityMgr,
		collector: metricsCollector,
		client:    backendClient,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start starts the agent
func (a *Agent) Start() error {
	log.Println("=== Smart Monitor Agent ===")
	log.Printf("Version: %s", a.config.AgentVersion)
	log.Printf("Hostname: %s", a.config.Hostname)
	log.Printf("IP Address: %s", a.config.IPAddress)
	log.Printf("Backend: %s", a.config.BackendAddr)
	log.Printf("Metrics Interval: %v", a.config.MetricsInterval)

	// Setup graceful shutdown
	a.setupSignalHandler()

	// Connect to backend with retry
	if err := a.connectWithRetry(); err != nil {
		return fmt.Errorf("failed to connect to backend: %w", err)
	}
	defer a.client.Close()

	// Register or load credentials
	if err := a.client.LoadOrRegister(a.ctx); err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}

	// Start monitoring loop with auto-reconnect
	return a.runWithReconnect()
}

// Stop stops the agent gracefully
func (a *Agent) Stop() {
	log.Println("⚠ Stopping agent...")
	a.cancel()
	time.Sleep(time.Second) // Give time for cleanup
	log.Println("Agent stopped")
}

// setupSignalHandler sets up graceful shutdown on signals
func (a *Agent) setupSignalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		log.Printf("\n⚠ Received signal: %v", sig)
		a.Stop()
	}()
}

// connectWithRetry connects to backend with retry logic
func (a *Agent) connectWithRetry() error {
	for i := 0; i < a.config.MaxRetries; i++ {
		if i > 0 {
			log.Printf("Retry attempt %d/%d...", i+1, a.config.MaxRetries)
			time.Sleep(a.config.RetryInterval)
		}

		ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
		err := a.client.Connect(ctx)
		cancel()

		if err == nil {
			return nil
		}

		log.Printf("Connection failed: %v", err)
	}

	return fmt.Errorf("failed to connect after %d retries", a.config.MaxRetries)
}

// runWithReconnect runs the main loop with auto-reconnect
func (a *Agent) runWithReconnect() error {
	for {
		// Check if context is cancelled
		select {
		case <-a.ctx.Done():
			return nil
		default:
		}

		// Stream metrics
		err := a.client.StreamMetrics(a.ctx)

		// Check if it's a graceful shutdown
		if err == context.Canceled {
			return nil
		}

		if err != nil {
			log.Printf("Streaming error: %v", err)
			log.Printf("Reconnecting in %v...", a.config.ReconnectDelay)

			select {
			case <-a.ctx.Done():
				return nil
			case <-time.After(a.config.ReconnectDelay):
				// Try to reconnect
				if err := a.connectWithRetry(); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					continue
				}

				// Re-register if needed
				if err := a.client.LoadOrRegister(a.ctx); err != nil {
					log.Printf("Failed to register after reconnect: %v", err)
					continue
				}
			}
		}
	}
}
