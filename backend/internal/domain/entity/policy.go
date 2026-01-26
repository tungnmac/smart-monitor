// Package entity defines policy management
package entity

import (
	"time"
)

// Policy represents a monitoring policy with thresholds and actions
type Policy struct {
	PolicyID       string
	Name           string
	Description    string
	Thresholds     map[string]string // e.g., {"cpu": "80", "ram": "90"}
	Actions        []string          // e.g., ["alert", "restart", "email"]
	Metadata       map[string]string
	Enabled        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	AppliedAgents  []string // List of agent IDs this policy is applied to
	AllowedUserIDs []string // List of user IDs allowed to access/apply this policy
}

// PolicyStatus represents the state of a policy application
type PolicyStatus string

const (
	PolicyStatusActive   PolicyStatus = "active"
	PolicyStatusInactive PolicyStatus = "inactive"
	PolicyStatusPending  PolicyStatus = "pending"
)

// NewPolicy creates a new policy
func NewPolicy(policyID, name, description string, thresholds map[string]string, actions []string, metadata map[string]string) *Policy {
	now := time.Now()

	if thresholds == nil {
		thresholds = make(map[string]string)
	}
	if actions == nil {
		actions = []string{}
	}
	if metadata == nil {
		metadata = make(map[string]string)
	}

	return &Policy{
		PolicyID:       policyID,
		Name:           name,
		Description:    description,
		Thresholds:     thresholds,
		Actions:        actions,
		Metadata:       metadata,
		Enabled:        true,
		CreatedAt:      now,
		UpdatedAt:      now,
		AppliedAgents:  []string{},
		AllowedUserIDs: []string{},
	}
}

// Update updates policy fields
func (p *Policy) Update(name, description string, thresholds map[string]string, actions []string, metadata map[string]string) {
	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	if thresholds != nil {
		p.Thresholds = thresholds
	}
	if actions != nil {
		p.Actions = actions
	}
	if metadata != nil {
		p.Metadata = metadata
	}
	p.UpdatedAt = time.Now()
}

// Enable enables the policy
func (p *Policy) Enable() {
	p.Enabled = true
	p.UpdatedAt = time.Now()
}

// Disable disables the policy
func (p *Policy) Disable() {
	p.Enabled = false
	p.UpdatedAt = time.Now()
}

// ApplyToAgent adds agent to applied list
func (p *Policy) ApplyToAgent(agentID string) bool {
	// Check if already applied
	for _, id := range p.AppliedAgents {
		if id == agentID {
			return false // Already applied
		}
	}
	p.AppliedAgents = append(p.AppliedAgents, agentID)
	p.UpdatedAt = time.Now()
	return true
}

// UnapplyFromAgent removes agent from applied list
func (p *Policy) UnapplyFromAgent(agentID string) bool {
	for i, id := range p.AppliedAgents {
		if id == agentID {
			p.AppliedAgents = append(p.AppliedAgents[:i], p.AppliedAgents[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false // Not found
}

// IsAppliedTo checks if policy is applied to agent
func (p *Policy) IsAppliedTo(agentID string) bool {
	for _, id := range p.AppliedAgents {
		if id == agentID {
			return true
		}
	}
	return false
}

// AddAllowedUser adds a user ID to the allowed list
func (p *Policy) AddAllowedUser(userID string) bool {
	for _, id := range p.AllowedUserIDs {
		if id == userID {
			return false
		}
	}
	p.AllowedUserIDs = append(p.AllowedUserIDs, userID)
	p.UpdatedAt = time.Now()
	return true
}

// RemoveAllowedUser removes a user ID from the allowed list
func (p *Policy) RemoveAllowedUser(userID string) bool {
	for i, id := range p.AllowedUserIDs {
		if id == userID {
			p.AllowedUserIDs = append(p.AllowedUserIDs[:i], p.AllowedUserIDs[i+1:]...)
			p.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// IsUserAllowed checks if a user ID is allowed for this policy
func (p *Policy) IsUserAllowed(userID string) bool {
	for _, id := range p.AllowedUserIDs {
		if id == userID {
			return true
		}
	}
	return false
}
