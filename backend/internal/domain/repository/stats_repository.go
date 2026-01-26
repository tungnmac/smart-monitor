// Package repository defines repository interfaces
package repository

import (
	"context"
	"smart-monitor/backend/internal/domain/entity"
)

// StatsRepository defines the interface for stats persistence
type StatsRepository interface {
	// Save stores stats
	Save(ctx context.Context, stats *entity.Stats) error

	// Get retrieves stats by hostname
	Get(ctx context.Context, hostname string) (*entity.Stats, error)

	// GetAll retrieves all stats
	GetAll(ctx context.Context) ([]*entity.Stats, error)

	// Delete removes stats for a hostname
	Delete(ctx context.Context, hostname string) error

	// GetActiveHosts returns list of active hostnames
	GetActiveHosts(ctx context.Context) ([]string, error)
}

// HostRepository defines the interface for host persistence
type HostRepository interface {
	// Create creates a new host
	Create(ctx context.Context, host *entity.Host) error

	// Get retrieves a host by hostname
	Get(ctx context.Context, hostname string) (*entity.Host, error)

	// GetAll retrieves all hosts
	GetAll(ctx context.Context) ([]*entity.Host, error)

	// Update updates a host
	Update(ctx context.Context, host *entity.Host) error

	// Delete removes a host
	Delete(ctx context.Context, hostname string) error
}

// AgentRegistryRepository defines the interface for agent registry persistence
type AgentRegistryRepository interface {
	// Register creates a new agent registry entry
	Register(ctx context.Context, agent *entity.AgentRegistry) error

	// GetByAgentID retrieves an agent by agent ID
	GetByAgentID(ctx context.Context, agentID string) (*entity.AgentRegistry, error)

	// GetByToken retrieves an agent by access token
	GetByToken(ctx context.Context, token string) (*entity.AgentRegistry, error)

	// Update updates an agent registry
	Update(ctx context.Context, agent *entity.AgentRegistry) error

	// Delete removes an agent registry
	Delete(ctx context.Context, agentID string) error

	// GetAll retrieves all registered agents
	GetAll(ctx context.Context) ([]*entity.AgentRegistry, error)

	// GetActive retrieves all active agents
	GetActive(ctx context.Context) ([]*entity.AgentRegistry, error)
}
