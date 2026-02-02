// Package usecase implements application use cases
package usecase

import (
	"context"
	"smart-monitor/backend/internal/application/dto"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
	"smart-monitor/backend/internal/domain/service"
)

// MonitorUseCase handles monitoring use cases
type MonitorUseCase struct {
	statsService   *service.StatsService
	controlService *service.AgentControlService
	agentRegistry  interface {
		GetAll(ctx context.Context) ([]*entity.AgentRegistry, error)
	}
}

// NewMonitorUseCase creates a new MonitorUseCase
func NewMonitorUseCase(statsService *service.StatsService, controlService *service.AgentControlService) *MonitorUseCase {
	return &MonitorUseCase{
		statsService:   statsService,
		controlService: controlService,
		agentRegistry:  nil, // Will be set separately
	}
}

// SetAgentRegistry sets the agent registry repository
func (uc *MonitorUseCase) SetAgentRegistry(repo interface {
	GetAll(ctx context.Context) ([]*entity.AgentRegistry, error)
}) {
	uc.agentRegistry = repo
}

// RecordStats records incoming stats from agent
func (uc *MonitorUseCase) RecordStats(ctx context.Context, req *dto.StatsRequest) error {
	// Convert DTO to entity
	stats := entity.NewStats(req.Hostname, req.AgentID, req.IPAddress, req.CPU, req.RAM, req.Disk)
	if req.Metadata != nil {
		stats.Metadata = req.Metadata
	}

	// Process processes if present
	if len(req.Processes) > 0 {
		var processes []*entity.Process
		for _, p := range req.Processes {
			processes = append(processes, entity.NewProcess(
				p.PID, p.Name, p.CPU, p.Memory, req.AgentID, req.Hostname, p.Command, p.Port,
			))
		}
		uc.controlService.UpdateAgentProcesses(req.AgentID, processes)
	}

	// Process through domain service
	return uc.statsService.ProcessStats(ctx, stats, req.AgentVersion)
}

// GetStats retrieves stats for a hostname
func (uc *MonitorUseCase) GetStats(ctx context.Context, hostname string) (*dto.StatsResponse, error) {
	stats, err := uc.statsService.GetStats(ctx, hostname)
	if err != nil {
		return nil, err
	}

	// Convert entity to DTO
	return &dto.StatsResponse{
		Hostname:     stats.Hostname,
		AgentID:      stats.AgentID,
		IPAddress:    stats.IPAddress,
		CPU:          stats.CPU,
		RAM:          stats.RAM,
		Disk:         stats.Disk,
		Timestamp:    stats.Timestamp,
		LastReceived: stats.LastReceived,
		Metadata:     stats.Metadata,
	}, nil
}

// GetAllStats retrieves all stats
func (uc *MonitorUseCase) GetAllStats(ctx context.Context) ([]*dto.StatsResponse, error) {
	statsList, err := uc.statsService.GetAllStats(ctx)
	if err != nil {
		return nil, err
	}

	// Convert entities to DTOs
	responses := make([]*dto.StatsResponse, len(statsList))
	for i, stats := range statsList {
		responses[i] = &dto.StatsResponse{
			Hostname:     stats.Hostname,
			AgentID:      stats.AgentID,
			IPAddress:    stats.IPAddress,
			CPU:          stats.CPU,
			RAM:          stats.RAM,
			Disk:         stats.Disk,
			Timestamp:    stats.Timestamp,
			LastReceived: stats.LastReceived,
			Metadata:     stats.Metadata,
		}
	}

	return responses, nil
}

// GetActiveHosts returns list of active hosts
func (uc *MonitorUseCase) GetActiveHosts(ctx context.Context) ([]string, error) {
	return uc.statsService.GetActiveHosts(ctx)
}

// ListAllAgents returns all registered agents
func (uc *MonitorUseCase) ListAllAgents(ctx context.Context) ([]*entity.AgentRegistry, error) {
	if uc.agentRegistry == nil {
		return []*entity.AgentRegistry{}, nil
	}
	return uc.agentRepo().GetAll(ctx)
}

// CleanupAgents removes duplicate/stale agents
func (uc *MonitorUseCase) CleanupAgents(ctx context.Context) error {
	repo := uc.agentRepo()
	if repo == nil {
		return nil
	}

	// 1. Get all agents before cleanup to know what we might be deleting
	allAgents, _ := repo.GetAll(ctx)

	// 2. Perform repository cleanup
	err := repo.Cleanup(ctx)
	if err != nil {
		return err
	}

	// 3. Get remaining agents
	remainingAgents, _ := repo.GetAll(ctx)
	remainingIDs := make(map[string]bool)
	for _, a := range remainingAgents {
		remainingIDs[a.AgentID] = true
	}

	// 4. Clear process cache for removed agents
	for _, a := range allAgents {
		if !remainingIDs[a.AgentID] {
			uc.controlService.ClearAgentCache(a.AgentID)
			uc.controlService.UnregisterCommandChan(a.AgentID)
		}
	}

	return nil
}

// helper to get agent repository with cleanup capabilities
func (uc *MonitorUseCase) agentRepo() repository.AgentRegistryRepository {
	if repo, ok := uc.agentRegistry.(repository.AgentRegistryRepository); ok {
		return repo
	}
	return nil
}
