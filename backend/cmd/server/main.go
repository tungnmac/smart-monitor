// Smart Monitor Backend Server with OpenSearch Integration
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smart-monitor/backend/internal/application/usecase"
	"smart-monitor/backend/internal/domain/service"
	grpchandler "smart-monitor/backend/internal/infrastructure/grpc"
	httphandler "smart-monitor/backend/internal/infrastructure/http"
	"smart-monitor/backend/internal/infrastructure/opensearch"
	"smart-monitor/backend/internal/infrastructure/persistence"
	"smart-monitor/backend/pkg/config"
	pb "smart-monitor/pbtypes/monitor"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("=== Smart Monitor Backend (DDD Architecture + OpenSearch) ===")

	// Load configuration
	cfg := config.Load()
	log.Printf("Configuration loaded: gRPC=:%s, HTTP=:%s", cfg.Server.GRPCPort, cfg.Server.HTTPPort)

	// Initialize in-memory repositories as fallback
	statsRepo := persistence.NewInMemoryStatsRepository()
	hostRepo := persistence.NewInMemoryHostRepository()
	agentRepo := persistence.NewInMemoryAgentRegistryRepository()
	policyRepo := persistence.NewInMemoryPolicyRepository()
	userRepo := persistence.NewInMemoryUserRepository()
	log.Println("✓ In-memory repositories initialized (fallback)")

	// Initialize OpenSearch
	var osClient *opensearch.Client
	var osStatsRepo *opensearch.OpenSearchStatsRepository
	var osAlertsRepo *opensearch.AlertsRepository
	var osEventsRepo *opensearch.EventsRepository

	osConfig := config.LoadOpenSearchConfig()
	osClient, err := opensearch.NewClient(osConfig.Host, osConfig.Port, osConfig.Username, osConfig.Password, osConfig.InsecureSkipVerify)
	if err != nil {
		log.Printf("⚠ OpenSearch connection failed: %v (continuing with in-memory storage)", err)
	} else {
		defer osClient.Close()
		log.Println("✓ OpenSearch client initialized")

		// Initialize indexes
		if err := opensearch.InitializeIndexes(osClient); err != nil {
			log.Printf("⚠ Failed to initialize OpenSearch indexes: %v", err)
		}

		// Create OpenSearch repositories
		osStatsRepoTemp, err := opensearch.NewOpenSearchStatsRepository(osClient)
		if err == nil {
			osStatsRepo = osStatsRepoTemp
			statsRepo = osStatsRepo
			log.Println("✓ Using OpenSearch for stats storage")
		}

		osAlertsRepoTemp, err := opensearch.NewAlertsRepository(osClient)
		if err == nil {
			osAlertsRepo = osAlertsRepoTemp
			log.Println("✓ Alerts repository initialized")
		}

		osEventsRepoTemp, err := opensearch.NewEventsRepository(osClient)
		if err == nil {
			osEventsRepo = osEventsRepoTemp
			log.Println("✓ Events repository initialized")
		}
	}

	// Initialize domain services
	statsService := service.NewStatsService(statsRepo, hostRepo)
	authService := service.NewAuthService(agentRepo)
	controlService := service.NewAgentControlService(agentRepo)
	policyService := service.NewPolicyService(policyRepo, agentRepo)
	log.Println("✓ Domain services initialized")

	// Initialize use cases
	monitorUseCase := usecase.NewMonitorUseCase(statsService)
	log.Println("✓ Use cases initialized")

	// Initialize user auth service
	authCfg := config.LoadAuthConfig()
	userAuthService := service.NewUserAuthService(userRepo, authCfg.JWTSecret)
	log.Println("✓ User auth service initialized")

	// Initialize gRPC handlers
	monitorGRPCHandler := grpchandler.NewMonitorServiceServer(monitorUseCase, authService, controlService, policyService)
	log.Println("✓ gRPC handlers initialized")

	// Start gRPC server
	grpcServer, lis := startGRPCServer(cfg, monitorGRPCHandler)
	log.Printf("✓ gRPC Server starting on port :%s", cfg.Server.GRPCPort)

	// Start HTTP server
	httpServer := startHTTPServer(cfg, monitorUseCase, osClient, osStatsRepo, osAlertsRepo, osEventsRepo, userAuthService, policyService)
	log.Printf("✓ HTTP Gateway starting on port :%s", cfg.Server.HTTPPort)
	log.Printf("  → API:     http://localhost:%s/v1/", cfg.Server.HTTPPort)
	log.Printf("  → Swagger: http://localhost:%s/swagger/", cfg.Server.HTTPPort)
	log.Printf("  → Health:  http://localhost:%s/health", cfg.Server.HTTPPort)
	if osClient != nil {
		log.Printf("  → Search:  http://localhost:%s/search/", cfg.Server.HTTPPort)
	}

	// Wait a moment for gRPC server to start
	time.Sleep(100 * time.Millisecond)
	log.Println("=== Smart Monitor Backend Ready ===")

	// Setup graceful shutdown
	gracefulShutdown(grpcServer, httpServer, lis)
}

// startGRPCServer starts the gRPC server
func startGRPCServer(cfg *config.Config, monitorHandler *grpchandler.MonitorServiceServer) (*grpc.Server, net.Listener) {
	lis, err := net.Listen("tcp", ":"+cfg.Server.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.Server.GRPCPort, err)
	}

	// Create gRPC server with health check
	grpcServer := grpc.NewServer()

	// Register services
	pb.RegisterMonitorServiceServer(grpcServer, monitorHandler)

	// Register health check service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	// Start server in goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	return grpcServer, lis
}

// startHTTPServer starts the HTTP gateway server
func startHTTPServer(cfg *config.Config, monitorUseCase *usecase.MonitorUseCase, osClient *opensearch.Client, osStatsRepo *opensearch.OpenSearchStatsRepository, osAlertsRepo *opensearch.AlertsRepository, osEventsRepo *opensearch.EventsRepository, userAuthService *service.UserAuthService, policyService *service.PolicyService) *http.Server {
	ctx := context.Background()

	// Create HTTP mux
	httpMux := http.NewServeMux()

	// Create gRPC gateway mux
	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := fmt.Sprintf("localhost:%s", cfg.Server.GRPCPort)

	err := pb.RegisterMonitorServiceHandlerFromEndpoint(ctx, gwMux, endpoint, opts)
	if err != nil {
		log.Fatalf("Failed to register gateway: %v", err)
	}

	// Mount API gateway
	httpMux.Handle("/v1/", gwMux)

	// Health check endpoints
	httpMux.Handle("/health", httphandler.NewHealthHandler(monitorUseCase))
	httpMux.Handle("/ready", httphandler.NewReadyHandler(monitorUseCase))
	httpMux.Handle("/live", httphandler.NewLiveHandler())
	httpMux.Handle("/metrics", httphandler.NewMetricsHandler(monitorUseCase))

	// Auth endpoints
	authHandler := httphandler.NewAuthHandler(userAuthService)
	httpMux.HandleFunc("/auth/signup", authHandler.SignUp)
	httpMux.HandleFunc("/auth/signin", authHandler.SignIn)

	// Admin tools: create user with custom role (no auth middleware per request)
	adminUserHandler := httphandler.NewAdminUserHandler(userAuthService)
	httpMux.HandleFunc("/tools/users", adminUserHandler.AddUser)

	// Search and storage endpoints (if OpenSearch is available)
	if osClient != nil && osStatsRepo != nil && osAlertsRepo != nil && osEventsRepo != nil {
		searchHandler := httphandler.NewSearchHandler(osStatsRepo, osAlertsRepo, osEventsRepo)

		// Read-only endpoints (all roles)
		httpMux.HandleFunc("/search/stats", searchHandler.SearchStats)
		httpMux.HandleFunc("/search/alerts", searchHandler.SearchAlerts)
		httpMux.HandleFunc("/search/events", searchHandler.SearchEvents)
		httpMux.HandleFunc("/search/alerts/stats", searchHandler.GetAlertStats)
		httpMux.HandleFunc("/search/events/stats", searchHandler.GetEventStats)

		// Write endpoints require admin or operator
		httpMux.HandleFunc("/search/alerts/create", httphandler.RequireRoles(userAuthService, []string{"admin", "operator"}, searchHandler.CreateAlert))
		httpMux.HandleFunc("/search/alerts/resolve", httphandler.RequireRoles(userAuthService, []string{"admin", "operator"}, searchHandler.ResolveAlert))
		httpMux.HandleFunc("/search/events/log", httphandler.RequireRoles(userAuthService, []string{"admin", "operator"}, searchHandler.LogEvent))

		log.Println("✓ Search endpoints registered (with RBAC)")
	}

	// Policy access management endpoints (RBAC protected)
	policyAccessHandler := httphandler.NewPolicyAccessHandler(policyService, userAuthService)
	httpMux.HandleFunc("/v1/policies/", httphandler.RequireRoles(userAuthService, []string{"admin", "operator"}, policyAccessHandler.ServeHTTP))

	// Swagger endpoints - Dynamic API documentation
	// Main Swagger JSON endpoint
	httpMux.HandleFunc("/v1/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./static/swagger.json")
	})

	// Alternative Swagger JSON endpoints for compatibility
	httpMux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./static/swagger.json")
	})
	httpMux.HandleFunc("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./static/swagger.json")
	})

	// Swagger UI endpoints
	httpMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		// Serve custom swagger UI HTML
		if r.URL.Path == "/swagger/" || r.URL.Path == "/swagger" {
			http.ServeFile(w, r, "./static/swagger-ui.html")
			return
		}
		// Serve other swagger assets
		http.StripPrefix("/swagger/", http.FileServer(http.Dir("./static/"))).ServeHTTP(w, r)
	})

	// Legacy endpoint for backward compatibility
	httpMux.HandleFunc("/v1/swagger/monitor.swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./static/swagger.json")
	})

	// Root endpoint
	httpMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":      "Smart Monitor Backend",
			"version":      "1.0.0",
			"architecture": "DDD (Domain-Driven Design) + OpenSearch",
			"endpoints": map[string]string{
				"health":  "/health",
				"ready":   "/ready",
				"live":    "/live",
				"metrics": "/metrics",
				"api":     "/v1/",
				"swagger": "/swagger/",
				"search":  "/search/",
			},
		})
	})

	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.HTTPPort,
		Handler:      httpMux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	return httpServer
}

// gracefulShutdown handles graceful shutdown
func gracefulShutdown(grpcServer *grpc.Server, httpServer *http.Server, lis net.Listener) {
	// Setup signal handling
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

	// Close listener
	if err := lis.Close(); err != nil {
		log.Printf("Listener close error: %v", err)
	}

	log.Println("=== Smart Monitor Backend Stopped ===")
}
