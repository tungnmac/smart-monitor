// Package dto defines data transfer objects
package dto

import "time"

// StatsRequest represents incoming stats request
type StatsRequest struct {
	Hostname     string
	AgentID      string
	IPAddress    string
	AgentVersion string
	CPU          float64
	RAM          float64
	Disk         float64
	Metadata     map[string]string
}

// StatsResponse represents stats response
type StatsResponse struct {
	Hostname     string
	AgentID      string
	IPAddress    string
	CPU          float64
	RAM          float64
	Disk         float64
	Timestamp    time.Time
	LastReceived time.Time
	Metadata     map[string]string
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp int64                  `json:"timestamp"`
	Service   string                 `json:"service"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ReadyResponse represents readiness check response
type ReadyResponse struct {
	Status      string   `json:"status"`
	Timestamp   int64    `json:"timestamp"`
	ActiveHosts []string `json:"active_hosts"`
}
