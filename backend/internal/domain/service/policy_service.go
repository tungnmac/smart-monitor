// Package service implements business logic
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
	"time"
)

// PolicyService handles policy business logic
type PolicyService struct {
	policyRepo repository.PolicyRepository
	agentRepo  repository.AgentRegistryRepository
}

// NewPolicyService creates a new policy service
func NewPolicyService(policyRepo repository.PolicyRepository, agentRepo repository.AgentRegistryRepository) *PolicyService {
	return &PolicyService{
		policyRepo: policyRepo,
		agentRepo:  agentRepo,
	}
}

// CreatePolicy creates a new policy
func (s *PolicyService) CreatePolicy(name, description string, thresholds map[string]string, actions []string, metadata map[string]string) (*entity.Policy, error) {
	// Generate policy ID
	policyID := s.generatePolicyID(name)

	policy := entity.NewPolicy(policyID, name, description, thresholds, actions, metadata)

	if err := s.policyRepo.Create(policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// UpdatePolicy updates an existing policy
func (s *PolicyService) UpdatePolicy(policyID, name, description string, thresholds map[string]string, actions []string, metadata map[string]string) (*entity.Policy, error) {
	policy, err := s.policyRepo.GetByID(policyID)
	if err != nil {
		return nil, err
	}

	policy.Update(name, description, thresholds, actions, metadata)

	if err := s.policyRepo.Update(policy); err != nil {
		return nil, err
	}

	return policy, nil
}

// RemovePolicy deletes a policy
func (s *PolicyService) RemovePolicy(policyID string) error {
	return s.policyRepo.Delete(policyID)
}

// GetPolicy retrieves a policy by ID
func (s *PolicyService) GetPolicy(policyID string) (*entity.Policy, error) {
	return s.policyRepo.GetByID(policyID)
}

// ListPolicies retrieves all policies with pagination
func (s *PolicyService) ListPolicies(page, pageSize int) ([]*entity.Policy, int, error) {
	return s.policyRepo.GetAll(page, pageSize)
}

// ApplyPolicyToAgent applies a policy to an agent
func (s *PolicyService) ApplyPolicyToAgent(policyID, agentID string) error {
	// Check if agent exists
	agent, err := s.agentRepo.GetByAgentID(context.Background(), agentID)
	if err != nil {
		return errors.New("agent not found")
	}

	if agent.Status != entity.AgentStatusActive {
		return errors.New("agent is not active")
	}

	// Check if policy exists
	_, err = s.policyRepo.GetByID(policyID)
	if err != nil {
		return errors.New("policy not found")
	}

	return s.policyRepo.ApplyToAgent(policyID, agentID)
}

// UnapplyPolicyFromAgent removes policy from agent
func (s *PolicyService) UnapplyPolicyFromAgent(policyID, agentID string) error {
	return s.policyRepo.UnapplyFromAgent(policyID, agentID)
}

// GetPoliciesByAgent retrieves policies applied to an agent
func (s *PolicyService) GetPoliciesByAgent(agentID string) ([]*entity.Policy, error) {
	return s.policyRepo.GetByAgent(agentID)
}

// EnablePolicy enables a policy
func (s *PolicyService) EnablePolicy(policyID string) error {
	policy, err := s.policyRepo.GetByID(policyID)
	if err != nil {
		return err
	}

	policy.Enable()
	return s.policyRepo.Update(policy)
}

// DisablePolicy disables a policy
func (s *PolicyService) DisablePolicy(policyID string) error {
	policy, err := s.policyRepo.GetByID(policyID)
	if err != nil {
		return err
	}

	policy.Disable()
	return s.policyRepo.Update(policy)
}

// AddAllowedUserToPolicy allows a specific user to access/apply a policy
func (s *PolicyService) AddAllowedUserToPolicy(policyID, userID string) error {
	policy, err := s.policyRepo.GetByID(policyID)
	if err != nil {
		return err
	}
	if !policy.AddAllowedUser(userID) {
		return errors.New("user already allowed")
	}
	return s.policyRepo.Update(policy)
}

// RemoveAllowedUserFromPolicy revokes a user's access to a policy
func (s *PolicyService) RemoveAllowedUserFromPolicy(policyID, userID string) error {
	policy, err := s.policyRepo.GetByID(policyID)
	if err != nil {
		return err
	}
	if !policy.RemoveAllowedUser(userID) {
		return errors.New("user not allowed")
	}
	return s.policyRepo.Update(policy)
}

// ListAllowedUsers returns user IDs allowed on a policy
func (s *PolicyService) ListAllowedUsers(policyID string) ([]string, error) {
	policy, err := s.policyRepo.GetByID(policyID)
	if err != nil {
		return nil, err
	}
	return append([]string{}, policy.AllowedUserIDs...), nil
}

// generatePolicyID generates a unique policy ID
func (s *PolicyService) generatePolicyID(name string) string {
	data := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return "policy-" + hex.EncodeToString(hash[:])[:8]
}
