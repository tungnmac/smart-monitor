// Package persistence implements data access layer
package persistence

import (
	"errors"
	"smart-monitor/backend/internal/domain/entity"
	"sync"
)

// InMemoryPolicyRepository implements PolicyRepository with in-memory storage
type InMemoryPolicyRepository struct {
	mu       sync.RWMutex
	policies map[string]*entity.Policy
}

// NewInMemoryPolicyRepository creates a new in-memory policy repository
func NewInMemoryPolicyRepository() *InMemoryPolicyRepository {
	return &InMemoryPolicyRepository{
		policies: make(map[string]*entity.Policy),
	}
}

// Create adds a new policy
func (r *InMemoryPolicyRepository) Create(policy *entity.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.policies[policy.PolicyID]; exists {
		return errors.New("policy already exists")
	}

	r.policies[policy.PolicyID] = policy
	return nil
}

// Update modifies an existing policy
func (r *InMemoryPolicyRepository) Update(policy *entity.Policy) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.policies[policy.PolicyID]; !exists {
		return errors.New("policy not found")
	}

	r.policies[policy.PolicyID] = policy
	return nil
}

// Delete removes a policy
func (r *InMemoryPolicyRepository) Delete(policyID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.policies[policyID]; !exists {
		return errors.New("policy not found")
	}

	delete(r.policies, policyID)
	return nil
}

// GetByID retrieves a policy by ID
func (r *InMemoryPolicyRepository) GetByID(policyID string) (*entity.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	policy, exists := r.policies[policyID]
	if !exists {
		return nil, errors.New("policy not found")
	}

	return policy, nil
}

// GetAll retrieves all policies with pagination
func (r *InMemoryPolicyRepository) GetAll(page, pageSize int) ([]*entity.Policy, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var policies []*entity.Policy
	for _, policy := range r.policies {
		policies = append(policies, policy)
	}

	total := len(policies)

	// Simple pagination
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		return []*entity.Policy{}, total, nil
	}
	if end > total {
		end = total
	}

	return policies[start:end], total, nil
}

// GetByAgent retrieves policies applied to an agent
func (r *InMemoryPolicyRepository) GetByAgent(agentID string) ([]*entity.Policy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var policies []*entity.Policy
	for _, policy := range r.policies {
		if policy.IsAppliedTo(agentID) {
			policies = append(policies, policy)
		}
	}

	return policies, nil
}

// ApplyToAgent applies a policy to an agent
func (r *InMemoryPolicyRepository) ApplyToAgent(policyID, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	policy, exists := r.policies[policyID]
	if !exists {
		return errors.New("policy not found")
	}

	if !policy.ApplyToAgent(agentID) {
		return errors.New("policy already applied to agent")
	}

	return nil
}

// UnapplyFromAgent removes policy from agent
func (r *InMemoryPolicyRepository) UnapplyFromAgent(policyID, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	policy, exists := r.policies[policyID]
	if !exists {
		return errors.New("policy not found")
	}

	if !policy.UnapplyFromAgent(agentID) {
		return errors.New("policy not applied to agent")
	}

	return nil
}
