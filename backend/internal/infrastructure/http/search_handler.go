// Package http provides HTTP handlers for search and storage
package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"smart-monitor/backend/internal/infrastructure/opensearch"
)

// SearchHandler handles search requests
type SearchHandler struct {
	statsRepo  *opensearch.OpenSearchStatsRepository
	alertsRepo *opensearch.AlertsRepository
	eventsRepo *opensearch.EventsRepository
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(statsRepo *opensearch.OpenSearchStatsRepository, alertsRepo *opensearch.AlertsRepository, eventsRepo *opensearch.EventsRepository) *SearchHandler {
	return &SearchHandler{
		statsRepo:  statsRepo,
		alertsRepo: alertsRepo,
		eventsRepo: eventsRepo,
	}
}

// SearchStats searches stats data
func (h *SearchHandler) SearchStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	hostname := r.URL.Query().Get("hostname")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	stats, err := h.statsRepo.SearchStats(r.Context(), query, hostname, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":  len(stats),
		"limit":  limit,
		"query":  query,
		"result": stats,
	})
}

// SearchAlerts searches alert data
func (h *SearchHandler) SearchAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hostname := r.URL.Query().Get("hostname")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	alerts, err := h.alertsRepo.ListAlerts(r.Context(), hostname, severity, status, limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":    len(alerts),
		"limit":    limit,
		"hostname": hostname,
		"severity": severity,
		"status":   status,
		"result":   alerts,
	})
}

// SearchEvents searches event data
func (h *SearchHandler) SearchEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	hostname := r.URL.Query().Get("hostname")
	eventType := r.URL.Query().Get("type")
	level := r.URL.Query().Get("level")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	events, err := h.eventsRepo.ListEvents(r.Context(), hostname, eventType, level, limit, offset)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":    len(events),
		"limit":    limit,
		"offset":   offset,
		"query":    query,
		"hostname": hostname,
		"type":     eventType,
		"level":    level,
		"result":   events,
	})
}

// GetAlertStats returns alert statistics
func (h *SearchHandler) GetAlertStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.alertsRepo.GetAlertStats(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetEventStats returns event statistics
func (h *SearchHandler) GetEventStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.eventsRepo.GetEventStats(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// CreateAlert creates a new alert
func (h *SearchHandler) CreateAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var alert opensearch.Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	id, err := h.alertsRepo.CreateAlert(r.Context(), &alert)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

// ResolveAlert marks an alert as resolved
func (h *SearchHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	alertID := r.URL.Query().Get("id")
	if alertID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Alert ID is required",
		})
		return
	}

	if err := h.alertsRepo.ResolveAlert(r.Context(), alertID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Alert resolved successfully",
	})
}

// LogEvent creates a new event
func (h *SearchHandler) LogEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event opensearch.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	id, err := h.eventsRepo.LogEvent(r.Context(), &event)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}
