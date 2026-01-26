// Package usecase implements application use cases
package usecase

import (
	"context"
	"smart-monitor/backend/internal/application/dto"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/service"
)

// MonitorUseCase handles monitoring use cases
type MonitorUseCase struct {
	statsService *service.StatsService
}

// NewMonitorUseCase creates a new MonitorUseCase
func NewMonitorUseCase(statsService *service.StatsService) *MonitorUseCase {
	return &MonitorUseCase{
		statsService: statsService,
	}
}

// RecordStats records incoming stats from agent
func (uc *MonitorUseCase) RecordStats(ctx context.Context, req *dto.StatsRequest) error {
	// Convert DTO to entity
	stats := entity.NewStats(req.Hostname, req.AgentID, req.IPAddress, req.CPU, req.RAM, req.Disk)
	if req.Metadata != nil {
		stats.Metadata = req.Metadata
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
