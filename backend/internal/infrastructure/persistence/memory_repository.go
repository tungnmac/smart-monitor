// Package persistence implements repository interfaces with in-memory storage
package persistence

import (
	"context"
	"fmt"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
	"sync"
)

// InMemoryStatsRepository implements StatsRepository with in-memory storage
type InMemoryStatsRepository struct {
	mu    sync.RWMutex
	stats map[string]*entity.Stats
}

// NewInMemoryStatsRepository creates a new in-memory stats repository
func NewInMemoryStatsRepository() repository.StatsRepository {
	return &InMemoryStatsRepository{
		stats: make(map[string]*entity.Stats),
	}
}

// Save stores stats
func (r *InMemoryStatsRepository) Save(ctx context.Context, stats *entity.Stats) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats[stats.Hostname] = stats
	return nil
}

// Get retrieves stats by hostname
func (r *InMemoryStatsRepository) Get(ctx context.Context, hostname string) (*entity.Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats, exists := r.stats[hostname]
	if !exists {
		return nil, fmt.Errorf("stats not found for hostname: %s", hostname)
	}

	return stats, nil
}

// GetAll retrieves all stats
func (r *InMemoryStatsRepository) GetAll(ctx context.Context) ([]*entity.Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entity.Stats, 0, len(r.stats))
	for _, stats := range r.stats {
		result = append(result, stats)
	}

	return result, nil
}

// Delete removes stats for a hostname
func (r *InMemoryStatsRepository) Delete(ctx context.Context, hostname string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.stats, hostname)
	return nil
}

// GetActiveHosts returns list of active hostnames
func (r *InMemoryStatsRepository) GetActiveHosts(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	hosts := make([]string, 0, len(r.stats))
	for hostname := range r.stats {
		hosts = append(hosts, hostname)
	}

	return hosts, nil
}

// InMemoryHostRepository implements HostRepository with in-memory storage
type InMemoryHostRepository struct {
	mu    sync.RWMutex
	hosts map[string]*entity.Host
}

// NewInMemoryHostRepository creates a new in-memory host repository
func NewInMemoryHostRepository() repository.HostRepository {
	return &InMemoryHostRepository{
		hosts: make(map[string]*entity.Host),
	}
}

// Create creates a new host
func (r *InMemoryHostRepository) Create(ctx context.Context, host *entity.Host) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hosts[host.Hostname] = host
	return nil
}

// Get retrieves a host by hostname
func (r *InMemoryHostRepository) Get(ctx context.Context, hostname string) (*entity.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	host, exists := r.hosts[hostname]
	if !exists {
		return nil, fmt.Errorf("host not found: %s", hostname)
	}

	return host, nil
}

// GetAll retrieves all hosts
func (r *InMemoryHostRepository) GetAll(ctx context.Context) ([]*entity.Host, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*entity.Host, 0, len(r.hosts))
	for _, host := range r.hosts {
		result = append(result, host)
	}

	return result, nil
}

// Update updates a host
func (r *InMemoryHostRepository) Update(ctx context.Context, host *entity.Host) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hosts[host.Hostname] = host
	return nil
}

// Delete removes a host
func (r *InMemoryHostRepository) Delete(ctx context.Context, hostname string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.hosts, hostname)
	return nil
}

// InMemoryAgentRegistryRepository implements AgentRegistryRepository with in-memory storage
type InMemoryAgentRegistryRepository struct {
	mu            sync.RWMutex
	agents        map[string]*entity.AgentRegistry // key: agentID
	tokenIndex    map[string]string                // key: token, value: agentID
	hostnameIndex map[string]string                // key: hostname, value: agentID
}

// NewInMemoryAgentRegistryRepository creates a new in-memory agent registry repository
func NewInMemoryAgentRegistryRepository() repository.AgentRegistryRepository {
	return &InMemoryAgentRegistryRepository{
		agents:        make(map[string]*entity.AgentRegistry),
		tokenIndex:    make(map[string]string),
		hostnameIndex: make(map[string]string),
	}
}

// Register creates a new agent registry entry
func (r *InMemoryAgentRegistryRepository) Register(ctx context.Context, agent *entity.AgentRegistry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Proactively clean up ANY existing agents with the same hostname to prevent duplicates
	for id, a := range r.agents {
		if a.Hostname == agent.Hostname && id != agent.AgentID {
			delete(r.tokenIndex, a.AccessToken)
			delete(r.agents, id)
		}
	}

	r.agents[agent.AgentID] = agent
	r.tokenIndex[agent.AccessToken] = agent.AgentID
	r.hostnameIndex[agent.Hostname] = agent.AgentID
	return nil
}

// GetByAgentID retrieves an agent by agent ID
func (r *InMemoryAgentRegistryRepository) GetByAgentID(ctx context.Context, agentID string) (*entity.AgentRegistry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// GetByHostname retrieves an agent by hostname
func (r *InMemoryAgentRegistryRepository) GetByHostname(ctx context.Context, hostname string) (*entity.AgentRegistry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentID, exists := r.hostnameIndex[hostname]
	if !exists {
		return nil, fmt.Errorf("agent not found for hostname: %s", hostname)
	}

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// GetByToken retrieves an agent by access token
func (r *InMemoryAgentRegistryRepository) GetByToken(ctx context.Context, token string) (*entity.AgentRegistry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentID, exists := r.tokenIndex[token]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	agent, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found for token")
	}

	return agent, nil
}

// Update updates an agent registry
func (r *InMemoryAgentRegistryRepository) Update(ctx context.Context, agent *entity.AgentRegistry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update indexes if needed
	if oldAgent, exists := r.agents[agent.AgentID]; exists {
		if oldAgent.AccessToken != agent.AccessToken {
			delete(r.tokenIndex, oldAgent.AccessToken)
			r.tokenIndex[agent.AccessToken] = agent.AgentID
		}
		if oldAgent.Hostname != agent.Hostname {
			delete(r.hostnameIndex, oldAgent.Hostname)
			r.hostnameIndex[agent.Hostname] = agent.AgentID
		}
	}

	r.agents[agent.AgentID] = agent
	r.tokenIndex[agent.AccessToken] = agent.AgentID
	r.hostnameIndex[agent.Hostname] = agent.AgentID
	return nil
}

// Delete removes an agent registry
func (r *InMemoryAgentRegistryRepository) Delete(ctx context.Context, agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent, exists := r.agents[agentID]; exists {
		delete(r.tokenIndex, agent.AccessToken)
		delete(r.hostnameIndex, agent.Hostname)
	}
	delete(r.agents, agentID)
	return nil
}

// GetAll retrieves all registered agents, deduplicated by hostname
func (r *InMemoryAgentRegistryRepository) GetAll(ctx context.Context) ([]*entity.AgentRegistry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Use a map to keep only the most recent agent for each hostname
	dedup := make(map[string]*entity.AgentRegistry)
	for _, agent := range r.agents {
		existing, exists := dedup[agent.Hostname]
		if !exists || agent.LastAuthAt.After(existing.LastAuthAt) {
			dedup[agent.Hostname] = agent
		}
	}

	result := make([]*entity.AgentRegistry, 0, len(dedup))
	for _, agent := range dedup {
		result = append(result, agent)
	}

	return result, nil
}

// GetActive retrieves all active agents, deduplicated by hostname
func (r *InMemoryAgentRegistryRepository) GetActive(ctx context.Context) ([]*entity.AgentRegistry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dedup := make(map[string]*entity.AgentRegistry)
	for _, agent := range r.agents {
		if agent.IsValid() {
			existing, exists := dedup[agent.Hostname]
			if !exists || agent.LastAuthAt.After(existing.LastAuthAt) {
				dedup[agent.Hostname] = agent
			}
		}
	}

	result := make([]*entity.AgentRegistry, 0, len(dedup))
	for _, agent := range dedup {
		result = append(result, agent)
	}

	return result, nil
}

// Cleanup removes duplicate agent entries, keeping only the most recent for each hostname
func (r *InMemoryAgentRegistryRepository) Cleanup(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 1. Identify best agent for each hostname (the one most recientemente seen)
	bestAgents := make(map[string]string) // hostname -> agentID
	for agentID, agent := range r.agents {
		if bestID, exists := bestAgents[agent.Hostname]; !exists {
			bestAgents[agent.Hostname] = agentID
		} else {
			bestAgent := r.agents[bestID]
			if agent.LastAuthAt.After(bestAgent.LastAuthAt) {
				bestAgents[agent.Hostname] = agentID
			}
		}
	}

	// 2. Remove all agents not in the bestAgents map
	toDelete := make([]string, 0)
	for agentID, agent := range r.agents {
		if bestID := bestAgents[agent.Hostname]; bestID != agentID {
			toDelete = append(toDelete, agentID)
		}
	}

	for _, agentID := range toDelete {
		agent := r.agents[agentID]
		delete(r.tokenIndex, agent.AccessToken)
		delete(r.agents, agentID)
	}

	// 3. Ensure hostnameIndex points to the remaining ones
	r.hostnameIndex = make(map[string]string)
	for id, agent := range r.agents {
		r.hostnameIndex[agent.Hostname] = id
	}

	return nil
}
