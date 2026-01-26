// Package entity defines core business entities
package entity

import "time"

// Stats represents system metrics for a host
type Stats struct {
	Hostname     string
	AgentID      string // Unique agent identifier
	IPAddress    string // Agent's IP address
	CPU          float64
	RAM          float64
	Disk         float64
	Timestamp    time.Time
	LastReceived time.Time
	Metadata     map[string]string // Additional metadata
}

// NewStats creates a new Stats instance
func NewStats(hostname, agentID, ipAddress string, cpu, ram, disk float64) *Stats {
	now := time.Now()
	return &Stats{
		Hostname:     hostname,
		AgentID:      agentID,
		IPAddress:    ipAddress,
		CPU:          cpu,
		RAM:          ram,
		Disk:         disk,
		Timestamp:    now,
		LastReceived: now,
		Metadata:     make(map[string]string),
	}
}

// IsValid checks if stats are valid
func (s *Stats) IsValid() bool {
	if s.Hostname == "" {
		return false
	}
	if s.AgentID == "" {
		return false
	}
	if s.CPU < 0 || s.CPU > 100 {
		return false
	}
	if s.RAM < 0 || s.RAM > 100 {
		return false
	}
	if s.Disk < 0 || s.Disk > 100 {
		return false
	}
	return true
}

// Update updates stats with new values
func (s *Stats) Update(cpu, ram, disk float64) {
	s.CPU = cpu
	s.RAM = ram
	s.Disk = disk
	s.Timestamp = time.Now()
	s.LastReceived = time.Now()
}
