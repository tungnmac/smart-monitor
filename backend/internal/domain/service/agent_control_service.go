// Package service implements control service for agents
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
	"sync"
)

// AgentControlService handles agent control operations
type AgentControlService struct {
	agentRepo    repository.AgentRegistryRepository
	processCache sync.Map // agentID -> []*entity.Process
	commandChans sync.Map // agentID -> chan *entity.AgentCommand
}

// NewAgentControlService creates a new agent control service
func NewAgentControlService(agentRepo repository.AgentRegistryRepository) *AgentControlService {
	return &AgentControlService{
		agentRepo:    agentRepo,
		commandChans: sync.Map{},
	}
}

// RegisterCommandChan registers a channel to send commands to an agent
func (s *AgentControlService) RegisterCommandChan(agentID string, ch chan *entity.AgentCommand) {
	s.commandChans.Store(agentID, ch)
}

// UnregisterCommandChan removes a command channel and optionally cleans up cache
func (s *AgentControlService) UnregisterCommandChan(agentID string) {
	s.commandChans.Delete(agentID)
}

// ClearAgentCache removes cached data for an agent
func (s *AgentControlService) ClearAgentCache(agentID string) {
	s.processCache.Delete(agentID)
}

// UpdateAgentProcesses updates the process list for an agent
func (s *AgentControlService) UpdateAgentProcesses(agentID string, processes []*entity.Process) {
	s.processCache.Store(agentID, processes)
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

// GetProcessList retrieves list of running processes from an agent
func (s *AgentControlService) GetProcessList(target string) ([]*entity.Process, error) {
	// Try to find agent by ID first
	agent, err := s.agentRepo.GetByAgentID(context.Background(), target)
	if err != nil {
		// Try by hostname if ID failed
		agent, err = s.agentRepo.GetByHostname(context.Background(), target)
		if err != nil {
			return nil, fmt.Errorf("agent not found (tried ID and hostname: %s)", target)
		}
	}

	agentID := agent.AgentID

	if agent.Status != entity.AgentStatusActive {
		return nil, fmt.Errorf("agent is not active (status: %s)", agent.Status)
	}

	if agent.IsBlocked() {
		return nil, errors.New("agent is blocked")
	}

	// Log the request
	log.Printf("Requesting process list from agent %s (host: %s)", agentID, agent.Hostname)

	// Check cache for real processes
	if val, ok := s.processCache.Load(agentID); ok {
		if processes, ok := val.([]*entity.Process); ok {
			log.Printf("✓ Retrieved %d real processes from cache for agent %s", len(processes), agentID)
			return processes, nil
		}
	}

	log.Printf("No real processes in cache for %s. Triggering refresh...", agentID)

	// Trigger an immediate heartbeat from the agent if possible
	if val, ok := s.commandChans.Load(agentID); ok {
		if ch, ok := val.(chan *entity.AgentCommand); ok {
			ch <- &entity.AgentCommand{Type: entity.CommandUpdateConfig}
			log.Printf("✓ Sent refresh command to agent %s", agentID)
		}
	}

	return nil, fmt.Errorf("no process data available for agent %s yet. A refresh request has been sent to the agent. Please try again in a few seconds", agentID)
}

// KillProcess sends a kill signal to a process on an agent
func (s *AgentControlService) KillProcess(target string, pid int32) error {
	// Try to find agent by ID first
	agent, err := s.agentRepo.GetByAgentID(context.Background(), target)
	if err != nil {
		// Try by hostname if ID failed
		agent, err = s.agentRepo.GetByHostname(context.Background(), target)
		if err != nil {
			return fmt.Errorf("agent not found (tried ID and hostname: %s)", target)
		}
	}

	agentID := agent.AgentID

	if agent.Status != entity.AgentStatusActive {
		return fmt.Errorf("agent is not active (status: %s)", agent.Status)
	}

	if agent.IsBlocked() {
		return errors.New("agent is blocked")
	}

	// Log the kill request
	log.Printf("Requesting to kill process %d on agent %s (host: %s)", pid, agentID, agent.Hostname)

	// Send command to the active stream
	if val, ok := s.commandChans.Load(agentID); ok {
		if ch, ok := val.(chan *entity.AgentCommand); ok {
			ch <- entity.NewKillProcessCommand(fmt.Sprintf("%d", pid))
			log.Printf("✓ Kill command queued for agent %s for PID %d", agentID, pid)
			return nil
		}
	}

	return fmt.Errorf("agent %s is connected but command channel is not available (is the stream active?)", agentID)
}
