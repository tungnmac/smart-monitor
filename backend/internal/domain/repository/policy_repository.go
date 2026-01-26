// Package repository defines data access interfaces
package repository

import "smart-monitor/backend/internal/domain/entity"

// PolicyRepository defines methods for policy persistence
type PolicyRepository interface {
	// Create adds a new policy
	Create(policy *entity.Policy) error

	// Update modifies an existing policy
	Update(policy *entity.Policy) error

	// Delete removes a policy
	Delete(policyID string) error

	// GetByID retrieves a policy by ID
	GetByID(policyID string) (*entity.Policy, error)

	// GetAll retrieves all policies with pagination
	GetAll(page, pageSize int) ([]*entity.Policy, int, error)

	// GetByAgent retrieves policies applied to an agent
	GetByAgent(agentID string) ([]*entity.Policy, error)

	// ApplyToAgent applies a policy to an agent
	ApplyToAgent(policyID, agentID string) error

	// UnapplyFromAgent removes policy from agent
	UnapplyFromAgent(policyID, agentID string) error
}
