// Package service defines domain services
package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
)

// AuthService handles agent authentication and registration
type AuthService struct {
	agentRepo repository.AgentRegistryRepository
}

// NewAuthService creates a new AuthService
func NewAuthService(agentRepo repository.AgentRegistryRepository) *AuthService {
	return &AuthService{
		agentRepo: agentRepo,
	}
}

// RegisterAgent registers a new agent and returns credentials
func (s *AuthService) RegisterAgent(ctx context.Context, hostname, ipAddress, agentVersion string, metadata map[string]string) (*entity.AgentRegistry, error) {
	// Generate unique agent ID
	agentID := generateAgentID(hostname, ipAddress)

	// Check if agent already exists
	existingAgent, err := s.agentRepo.GetByAgentID(ctx, agentID)
	if err == nil && existingAgent != nil {
		// Agent already registered, return existing credentials
		if existingAgent.IsValid() {
			existingAgent.UpdateLastAuth()
			if err := s.agentRepo.Update(ctx, existingAgent); err != nil {
				return nil, fmt.Errorf("failed to update existing agent: %w", err)
			}
			return existingAgent, nil
		}
		// Agent expired or suspended, renew token
		existingAgent.RenewToken()
		existingAgent.Activate()
		if err := s.agentRepo.Update(ctx, existingAgent); err != nil {
			return nil, fmt.Errorf("failed to renew agent token: %w", err)
		}
		return existingAgent, nil
	}

	// Create new agent registry
	agent := entity.NewAgentRegistry(agentID, hostname, ipAddress, agentVersion, metadata)

	if err := s.agentRepo.Register(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	return agent, nil
}

// ValidateToken validates an agent's access token
func (s *AuthService) ValidateToken(ctx context.Context, agentID, token string) error {
	agent, err := s.agentRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	if !agent.IsTokenValid(token) {
		return fmt.Errorf("invalid or expired token")
	}

	// Update last authentication time
	agent.UpdateLastAuth()
	if err := s.agentRepo.Update(ctx, agent); err != nil {
		return fmt.Errorf("failed to update auth timestamp: %w", err)
	}

	return nil
}

// GetAgentByToken retrieves agent by token
func (s *AuthService) GetAgentByToken(ctx context.Context, token string) (*entity.AgentRegistry, error) {
	agent, err := s.agentRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !agent.IsValid() {
		return nil, fmt.Errorf("agent is not active or token expired")
	}

	return agent, nil
}

// RevokeAgent revokes an agent's access
func (s *AuthService) RevokeAgent(ctx context.Context, agentID string) error {
	agent, err := s.agentRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	agent.Revoke()
	if err := s.agentRepo.Update(ctx, agent); err != nil {
		return fmt.Errorf("failed to revoke agent: %w", err)
	}

	return nil
}

// GetAllAgents retrieves all registered agents
func (s *AuthService) GetAllAgents(ctx context.Context) ([]*entity.AgentRegistry, error) {
	return s.agentRepo.GetAll(ctx)
}

// GetActiveAgents retrieves all active agents
func (s *AuthService) GetActiveAgents(ctx context.Context) ([]*entity.AgentRegistry, error) {
	return s.agentRepo.GetActive(ctx)
}

// generateAgentID creates a unique agent identifier
func generateAgentID(hostname, ipAddress string) string {
	data := fmt.Sprintf("%s-%s-%d", hostname, ipAddress, time.Now().Unix())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("agent-%x", hash[:8])
}
