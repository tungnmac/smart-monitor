package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "smart-monitor/pbtypes/monitor"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// Config holds server configuration
type Config struct {
	GRPCPort string
	HTTPPort string
}

// server implements the MonitorService
type server struct {
	pb.UnimplementedMonitorServiceServer
	mu           sync.RWMutex
	statsCache   map[string]*pb.StatsRequest
	lastReceived map[string]time.Time
}

// newServer creates a new server instance
func newServer() *server {
	return &server{
		statsCache:   make(map[string]*pb.StatsRequest),
		lastReceived: make(map[string]time.Time),
	}
}

// StreamStats handles bidirectional streaming from agents
func (s *server) StreamStats(stream pb.MonitorService_StreamStatsServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Stream closed by client")
			return stream.SendAndClose(&pb.StatsResponse{
				Message:   "Stream closed",
				Timestamp: time.Now().Unix(),
			})
		}
		if err != nil {
			log.Printf("Error receiving stats: %v", err)
			return fmt.Errorf("failed to receive stats: %w", err)
		}

		// Store stats in cache
		s.mu.Lock()
		s.statsCache[req.Hostname] = req
		s.lastReceived[req.Hostname] = time.Now()
		s.mu.Unlock()

		// Log received stats
		log.Printf("[%s] CPU: %.2f%% | RAM: %.2f%% | Disk: %.2f%%",
			req.Hostname, req.Cpu, req.Ram, req.Disk)

		// TODO: Store to database
		// TODO: Check alert thresholds
		// TODO: Push to WebSocket for real-time dashboard
	}
}

// GetStats returns stats for a specific hostname
func (s *server) GetStats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	hostname := req.Hostname
	if hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	log.Printf("GetStats called for hostname: %s", hostname)

	s.mu.RLock()
	stats, exists := s.statsCache[hostname]
	lastReceived := s.lastReceived[hostname]
	s.mu.RUnlock()

	if !exists {
		return &pb.StatsResponse{
			Message:   fmt.Sprintf("No stats available for hostname: %s", hostname),
			Timestamp: time.Now().Unix(),
		}, nil
	}

	message := fmt.Sprintf("Stats for %s: CPU=%.2f%%, RAM=%.2f%%, Disk=%.2f%% (Last received: %s)",
		hostname, stats.Cpu, stats.Ram, stats.Disk, lastReceived.Format(time.RFC3339))

	return &pb.StatsResponse{
		Message:   message,
		Timestamp: time.Now().Unix(),
	}, nil
}

// getActiveHosts returns list of active hosts
func (s *server) getActiveHosts() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hosts := make([]string, 0, len(s.statsCache))
	for host := range s.statsCache {
		hosts = append(hosts, host)
	}
	return hosts
}

// healthHandler returns health check status
func healthHandler(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if server is healthy
		// Could add database connection check here

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "smart-monitor-backend",
		})
	}
}

// readyHandler returns readiness status
func readyHandler(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "ready",
			"timestamp":    time.Now().Unix(),
			"active_hosts": s.getActiveHosts(),
		})
	}
}

// liveHandler returns liveness status
func liveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "alive",
			"timestamp": time.Now().Unix(),
		})
	}
}

// metricsHandler returns basic metrics (placeholder for Prometheus)
func metricsHandler(s *server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		hostCount := len(s.statsCache)
		s.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "# HELP smart_monitor_active_hosts Number of active hosts\n")
		fmt.Fprintf(w, "# TYPE smart_monitor_active_hosts gauge\n")
		fmt.Fprintf(w, "smart_monitor_active_hosts %d\n", hostCount)
	}
}

// getConfig returns configuration from environment variables or defaults
func getConfig() *Config {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	return &Config{
		GRPCPort: grpcPort,
		HTTPPort: httpPort,
	}
}

// startGRPCServer starts the gRPC server
func startGRPCServer(config *Config, srv *server) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", ":"+config.GRPCPort)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on port %s: %w", config.GRPCPort, err)
	}

	// Create gRPC server with health check
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterMonitorServiceServer(grpcServer, srv)

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	log.Printf("✓ gRPC Server starting on port :%s", config.GRPCPort)

	return grpcServer, lis, nil
}

// startHTTPServer starts the HTTP gateway server
func startHTTPServer(config *Config, srv *server) (*http.Server, error) {
	ctx := context.Background()

	// Create HTTP mux
	httpMux := http.NewServeMux()

	// Create gRPC gateway mux
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("localhost:%s", config.GRPCPort)

	err := pb.RegisterMonitorServiceHandlerFromEndpoint(ctx, gwMux, endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to register gateway: %w", err)
	}

	// Mount API gateway
	httpMux.Handle("/v1/", gwMux)

	// Health check endpoints
	httpMux.HandleFunc("/health", healthHandler(srv))
	httpMux.HandleFunc("/ready", readyHandler(srv))
	httpMux.HandleFunc("/live", liveHandler())
	httpMux.HandleFunc("/metrics", metricsHandler(srv))

	// Swagger endpoints
	httpMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../pbtypes/combined.swagger.json")
	})
	httpMux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./static/"))))

	// Root endpoint
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service": "Smart Monitor Backend",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"health":  "/health",
				"ready":   "/ready",
				"live":    "/live",
				"metrics": "/metrics",
				"api":     "/v1/",
				"swagger": "/swagger/",
			},
		})
	})

	httpServer := &http.Server{
		Addr:         ":" + config.HTTPPort,
		Handler:      httpMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("✓ HTTP Gateway starting on port :%s", config.HTTPPort)
	log.Printf("  → API:     http://localhost:%s/v1/", config.HTTPPort)
	log.Printf("  → Swagger: http://localhost:%s/swagger/", config.HTTPPort)
	log.Printf("  → Health:  http://localhost:%s/health", config.HTTPPort)

	return httpServer, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("=== Smart Monitor Backend Starting ===")

	// Load configuration
	config := getConfig()
	log.Printf("Configuration loaded: gRPC=:%s, HTTP=:%s", config.GRPCPort, config.HTTPPort)

	// Create server instance
	srv := newServer()

	// Start gRPC server
	grpcServer, lis, err := startGRPCServer(config, srv)
	if err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}

	// Start gRPC server in goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Wait a moment for gRPC server to start
	time.Sleep(100 * time.Millisecond)

	// Start HTTP server
	httpServer, err := startHTTPServer(config, srv)
	if err != nil {
		log.Fatalf("Failed to setup HTTP server: %v", err)
	}

	// Start HTTP server in goroutine
	go func() {
		log.Println("=== Smart Monitor Backend Ready ===")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	log.Println("\n=== Shutting down gracefully ===")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	log.Println("Stopping HTTP server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	log.Println("Stopping gRPC server...")
	grpcServer.GracefulStop()

	log.Println("=== Smart Monitor Backend Stopped ===")
}
