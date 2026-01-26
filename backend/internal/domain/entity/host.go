// Package entity defines core business entities
package entity

import "time"

// Host represents a monitored host/agent
type Host struct {
	ID           string
	Hostname     string
	IPAddress    string
	AgentID      string // Unique agent identifier
	AgentVersion string // Agent version
	Status       HostStatus
	Metadata     map[string]string // Additional metadata
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastSeenAt   time.Time
}

// HostStatus represents the status of a host
type HostStatus string

const (
	HostStatusOnline  HostStatus = "online"
	HostStatusOffline HostStatus = "offline"
	HostStatusUnknown HostStatus = "unknown"
)

// NewHost creates a new Host instance
func NewHost(hostname, ipAddress, agentID string) *Host {
	now := time.Now()
	return &Host{
		Hostname:   hostname,
		IPAddress:  ipAddress,
		AgentID:    agentID,
		Status:     HostStatusOnline,
		Metadata:   make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
		LastSeenAt: now,
	}
}

// UpdateStatus updates host status
func (h *Host) UpdateStatus(status HostStatus) {
	h.Status = status
	h.UpdatedAt = time.Now()
}

// MarkSeen updates last seen timestamp
func (h *Host) MarkSeen() {
	h.LastSeenAt = time.Now()
	h.UpdatedAt = time.Now()
	h.Status = HostStatusOnline
}

// IsActive checks if host has been seen recently (within 30 seconds)
func (h *Host) IsActive() bool {
	return time.Since(h.LastSeenAt) < 30*time.Second
}

// UpdateMetadata updates or adds metadata
func (h *Host) UpdateMetadata(key, value string) {
	if h.Metadata == nil {
		h.Metadata = make(map[string]string)
	}
	h.Metadata[key] = value
	h.UpdatedAt = time.Now()
}
