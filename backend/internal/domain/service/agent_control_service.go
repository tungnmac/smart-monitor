// Package service implements control service for agents
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
)

// AgentControlService handles agent control operations
type AgentControlService struct {
	agentRepo repository.AgentRegistryRepository
}

// NewAgentControlService creates a new agent control service
func NewAgentControlService(agentRepo repository.AgentRegistryRepository) *AgentControlService {
	return &AgentControlService{
		agentRepo: agentRepo,
	}
}

// ControlAgent sends control command to agent
func (s *AgentControlService) ControlAgent(agentID string, action entity.AgentControlAction, reason string) error {
	agent, err := s.agentRepo.GetByAgentID(context.Background(), agentID)
	if err != nil {
		return errors.New("agent not found")
	}

	if agent.Status != entity.AgentStatusActive {
		return fmt.Errorf("agent is not active (status: %s)", agent.Status)
	}

	if agent.IsBlocked() {
		return errors.New("agent is blocked")
	}

	// Log the control action
	log.Printf("Control action for agent %s: %s (reason: %s)", agentID, action, reason)

	// In a real implementation, this would send a message to the agent
	// For now, we just log it and return success
	switch action {
	case entity.AgentActionStart:
		log.Printf("Agent %s: Starting...", agentID)
	case entity.AgentActionShutdown:
		log.Printf("Agent %s: Shutting down...", agentID)
		agent.Suspend()
		s.agentRepo.Update(context.Background(), agent)
	case entity.AgentActionRestart:
		log.Printf("Agent %s: Restarting...", agentID)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

// BlockAgent blocks an agent
func (s *AgentControlService) BlockAgent(agentID string, reason string) error {
	agent, err := s.agentRepo.GetByAgentID(context.Background(), agentID)
	if err != nil {
		return errors.New("agent not found")
	}

	if agent.IsBlocked() {
		return errors.New("agent is already blocked")
	}

	agent.Block(reason)
	return s.agentRepo.Update(context.Background(), agent)
}

// UnblockAgent unblocks an agent
func (s *AgentControlService) UnblockAgent(agentID string) error {
	agent, err := s.agentRepo.GetByAgentID(context.Background(), agentID)
	if err != nil {
		return errors.New("agent not found")
	}

	if !agent.IsBlocked() {
		return errors.New("agent is not blocked")
	}

	agent.Unblock()
	return s.agentRepo.Update(context.Background(), agent)
}

// GetAgentStatus retrieves agent status
func (s *AgentControlService) GetAgentStatus(agentID string) (*entity.AgentRegistry, error) {
	return s.agentRepo.GetByAgentID(context.Background(), agentID)
}
