// Package entity defines core business entities
package entity

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// AgentRegistry represents a registered agent with authentication
type AgentRegistry struct {
	AgentID      string
	Hostname     string
	IPAddress    string
	AgentVersion string
	AccessToken  string
	TokenExpiry  time.Time
	Status       AgentStatus
	Blocked      bool
	BlockReason  string
	Metadata     map[string]string
	RegisteredAt time.Time
	LastAuthAt   time.Time
}

// AgentStatus represents agent registration status
type AgentStatus string

const (
	AgentStatusActive    AgentStatus = "active"
	AgentStatusSuspended AgentStatus = "suspended"
	AgentStatusRevoked   AgentStatus = "revoked"
	AgentStatusBlocked   AgentStatus = "blocked"
)

// AgentControlAction represents control actions for agents
type AgentControlAction string

const (
	AgentActionStart    AgentControlAction = "start"
	AgentActionShutdown AgentControlAction = "shutdown"
	AgentActionRestart  AgentControlAction = "restart"
)

// NewAgentRegistry creates a new agent registry entry
func NewAgentRegistry(agentID, hostname, ipAddress, agentVersion string, metadata map[string]string) *AgentRegistry {
	now := time.Now()
	token := generateAccessToken()

	return &AgentRegistry{
		AgentID:      agentID,
		Hostname:     hostname,
		IPAddress:    ipAddress,
		AgentVersion: agentVersion,
		AccessToken:  token,
		Blocked:      false,
		BlockReason:  "",
		TokenExpiry:  now.Add(365 * 24 * time.Hour), // 1 year validity
		Status:       AgentStatusActive,
		Metadata:     metadata,
		RegisteredAt: now,
		LastAuthAt:   now,
	}
}

// IsValid checks if agent registration is valid
func (a *AgentRegistry) IsValid() bool {
	return a.Status == AgentStatusActive && time.Now().Before(a.TokenExpiry)
}

// IsTokenValid checks if the provided token matches and is not expired
func (a *AgentRegistry) IsTokenValid(token string) bool {
	return a.AccessToken == token && a.IsValid()
}

// RenewToken generates a new access token
func (a *AgentRegistry) RenewToken() {
	a.AccessToken = generateAccessToken()
	a.TokenExpiry = time.Now().Add(365 * 24 * time.Hour)
	a.LastAuthAt = time.Now()
}

// Suspend suspends the agent
func (a *AgentRegistry) Suspend() {
	a.Status = AgentStatusSuspended
}

// Activate activates the agent
func (a *AgentRegistry) Activate() {
	a.Status = AgentStatusActive
	a.LastAuthAt = time.Now()
}

// Revoke revokes agent access permanently
func (a *AgentRegistry) Revoke() {
	a.Status = AgentStatusRevoked
	a.AccessToken = ""
}

// UpdateLastAuth updates the last authentication timestamp
func (a *AgentRegistry) UpdateLastAuth() {
	a.LastAuthAt = time.Now()
}

// generateAccessToken generates a secure random token
func generateAccessToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based token (less secure but still functional)
		return hex.EncodeToString([]byte(time.Now().String()))
	}
	return hex.EncodeToString(bytes)
}

// Block blocks the agent
func (a *AgentRegistry) Block(reason string) {
	a.Blocked = true
	a.BlockReason = reason
	a.Status = AgentStatusBlocked
}

// Unblock unblocks the agent
func (a *AgentRegistry) Unblock() {
	a.Blocked = false
	a.BlockReason = ""
	if a.Status == AgentStatusBlocked {
		a.Status = AgentStatusActive
	}
}

// IsBlocked checks if agent is blocked
func (a *AgentRegistry) IsBlocked() bool {
	return a.Blocked
}
