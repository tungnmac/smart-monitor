// Package service defines domain services
package service

import (
	"context"
	"fmt"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
)

// StatsService provides business logic for stats
type StatsService struct {
	statsRepo repository.StatsRepository
	hostRepo  repository.HostRepository
}

// NewStatsService creates a new StatsService
func NewStatsService(statsRepo repository.StatsRepository, hostRepo repository.HostRepository) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
		hostRepo:  hostRepo,
	}
}

// ProcessStats processes incoming stats from agent
func (s *StatsService) ProcessStats(ctx context.Context, stats *entity.Stats, agentVersion string) error {
	// Validate stats
	if !stats.IsValid() {
		return fmt.Errorf("invalid stats data from agent %s", stats.AgentID)
	}

	// Save stats
	if err := s.statsRepo.Save(ctx, stats); err != nil {
		return fmt.Errorf("failed to save stats: %w", err)
	}

	// Update or create host
	host, err := s.hostRepo.Get(ctx, stats.AgentID)
	if err != nil {
		// Create new host if not exists
		host = entity.NewHost(stats.Hostname, stats.IPAddress, stats.AgentID)
		host.AgentVersion = agentVersion
		if stats.Metadata != nil {
			host.Metadata = stats.Metadata
		}
		if err := s.hostRepo.Create(ctx, host); err != nil {
			return fmt.Errorf("failed to create host: %w", err)
		}
	} else {
		// Update existing host
		host.MarkSeen()
		host.Hostname = stats.Hostname   // Update hostname if changed
		host.IPAddress = stats.IPAddress // Update IP if changed
		host.AgentVersion = agentVersion // Update version
		if stats.Metadata != nil {
			for k, v := range stats.Metadata {
				host.UpdateMetadata(k, v)
			}
		}
		if err := s.hostRepo.Update(ctx, host); err != nil {
			return fmt.Errorf("failed to update host: %w", err)
		}
	}

	return nil
}

// GetStats retrieves stats for a hostname
func (s *StatsService) GetStats(ctx context.Context, hostname string) (*entity.Stats, error) {
	if hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	stats, err := s.statsRepo.Get(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}

// GetAllStats retrieves all stats
func (s *StatsService) GetAllStats(ctx context.Context) ([]*entity.Stats, error) {
	stats, err := s.statsRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all stats: %w", err)
	}

	return stats, nil
}

// GetActiveHosts returns list of active hosts
func (s *StatsService) GetActiveHosts(ctx context.Context) ([]string, error) {
	hosts, err := s.statsRepo.GetActiveHosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active hosts: %w", err)
	}

	return hosts, nil
}
