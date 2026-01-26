// Package http implements HTTP handlers
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"smart-monitor/backend/internal/application/usecase"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	monitorUseCase *usecase.MonitorUseCase
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(monitorUseCase *usecase.MonitorUseCase) *HealthHandler {
	return &HealthHandler{
		monitorUseCase: monitorUseCase,
	}
}

// ServeHTTP handles health check requests
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "smart-monitor-backend",
	})
}

// ReadyHandler handles readiness check requests
type ReadyHandler struct {
	monitorUseCase *usecase.MonitorUseCase
}

// NewReadyHandler creates a new ready handler
func NewReadyHandler(monitorUseCase *usecase.MonitorUseCase) *ReadyHandler {
	return &ReadyHandler{
		monitorUseCase: monitorUseCase,
	}
}

// ServeHTTP handles readiness check requests
func (h *ReadyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.monitorUseCase.GetActiveHosts(r.Context())
	if err != nil {
		hosts = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ready",
		"timestamp":    time.Now().Unix(),
		"active_hosts": hosts,
	})
}

// LiveHandler handles liveness check requests
type LiveHandler struct{}

// NewLiveHandler creates a new live handler
func NewLiveHandler() *LiveHandler {
	return &LiveHandler{}
}

// ServeHTTP handles liveness check requests
func (h *LiveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().Unix(),
	})
}

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	monitorUseCase *usecase.MonitorUseCase
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(monitorUseCase *usecase.MonitorUseCase) *MetricsHandler {
	return &MetricsHandler{
		monitorUseCase: monitorUseCase,
	}
}

// ServeHTTP handles metrics requests
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hosts, err := h.monitorUseCase.GetActiveHosts(r.Context())
	if err != nil {
		hosts = []string{}
	}

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP smart_monitor_active_hosts Number of active hosts\n")
	fmt.Fprintf(w, "# TYPE smart_monitor_active_hosts gauge\n")
	fmt.Fprintf(w, "smart_monitor_active_hosts %d\n", len(hosts))
}
