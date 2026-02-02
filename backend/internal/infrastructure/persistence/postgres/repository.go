package postgres

import (
	"context"
	"encoding/json"

	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresAgentRepository struct {
	db *gorm.DB
}

func NewPostgresAgentRepository(db *gorm.DB) repository.AgentRegistryRepository {
	return &PostgresAgentRepository{db: db}
}

func (r *PostgresAgentRepository) Register(ctx context.Context, agent *entity.AgentRegistry) error {
	metadata, _ := json.Marshal(agent.Metadata)
	model := AgentModel{
		AgentID:      agent.AgentID,
		Hostname:     agent.Hostname,
		IPAddress:    agent.IPAddress,
		AgentVersion: agent.AgentVersion,
		AccessToken:  agent.AccessToken,
		TokenExpiry:  agent.TokenExpiry,
		Status:       string(agent.Status),
		Blocked:      false,
		RegisteredAt: agent.RegisteredAt,
		LastAuthAt:   agent.LastAuthAt,
		Metadata:     string(metadata),
	}

	// Use Upsert logic to ensure uniqueness on Hostname
	// If Hostname exists, update all fields except RegisteredAt
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "hostname"}},
		DoUpdates: clause.AssignmentColumns([]string{"agent_id", "ip_address", "agent_version", "access_token", "token_expiry", "status", "last_auth_at", "metadata"}),
	}).Create(&model).Error
}

func (r *PostgresAgentRepository) GetByAgentID(ctx context.Context, agentID string) (*entity.AgentRegistry, error) {
	var model AgentModel
	if err := r.db.WithContext(ctx).First(&model, "agent_id = ?", agentID).Error; err != nil {
		return nil, err
	}
	return toAgentEntity(&model), nil
}

func (r *PostgresAgentRepository) GetByHostname(ctx context.Context, hostname string) (*entity.AgentRegistry, error) {
	var model AgentModel
	if err := r.db.WithContext(ctx).First(&model, "hostname = ?", hostname).Error; err != nil {
		return nil, err
	}
	return toAgentEntity(&model), nil
}

func (r *PostgresAgentRepository) GetByToken(ctx context.Context, token string) (*entity.AgentRegistry, error) {
	var model AgentModel
	if err := r.db.WithContext(ctx).First(&model, "access_token = ?", token).Error; err != nil {
		return nil, err
	}
	return toAgentEntity(&model), nil
}

func (r *PostgresAgentRepository) Update(ctx context.Context, agent *entity.AgentRegistry) error {
	metadata, _ := json.Marshal(agent.Metadata)
	updates := map[string]interface{}{
		"ip_address":    agent.IPAddress,
		"agent_version": agent.AgentVersion,
		"access_token":  agent.AccessToken,
		"token_expiry":  agent.TokenExpiry,
		"status":        string(agent.Status),
		"last_auth_at":  agent.LastAuthAt,
		"metadata":      string(metadata),
	}
	return r.db.WithContext(ctx).Model(&AgentModel{}).Where("agent_id = ?", agent.AgentID).Updates(updates).Error
}

func (r *PostgresAgentRepository) Delete(ctx context.Context, agentID string) error {
	return r.db.WithContext(ctx).Delete(&AgentModel{}, "agent_id = ?", agentID).Error
}

func (r *PostgresAgentRepository) GetAll(ctx context.Context) ([]*entity.AgentRegistry, error) {
	var models []AgentModel
	// Subquery to get latest agent per hostname
	subQuery := r.db.Model(&AgentModel{}).Select("MAX(last_auth_at) as max_auth, hostname").Group("hostname")
	if err := r.db.WithContext(ctx).Joins("JOIN (?) as latest ON latest.hostname = agent_models.hostname AND latest.max_auth = agent_models.last_auth_at", subQuery).Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*entity.AgentRegistry, len(models))
	for i, m := range models {
		result[i] = toAgentEntity(&m)
	}
	return result, nil
}

func (r *PostgresAgentRepository) GetActive(ctx context.Context) ([]*entity.AgentRegistry, error) {
	var models []AgentModel
	// Active status and deduplicated by hostname
	subQuery := r.db.Model(&AgentModel{}).Where("status = ?", string(entity.AgentStatusActive)).Select("MAX(last_auth_at) as max_auth, hostname").Group("hostname")
	if err := r.db.WithContext(ctx).Joins("JOIN (?) as latest ON latest.hostname = agent_models.hostname AND latest.max_auth = agent_models.last_auth_at", subQuery).Where("status = ?", string(entity.AgentStatusActive)).Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*entity.AgentRegistry, 0)
	for _, m := range models {
		agent := toAgentEntity(&m)
		if agent.IsValid() {
			result = append(result, agent)
		}
	}
	return result, nil
}

func (r *PostgresAgentRepository) Cleanup(ctx context.Context) error {
	// Remove older duplicates for the same hostname
	// We keep the one with the latest last_auth_at or registered_at
	// Using a more robust query that handles GORM default table naming (agent_models)
	query := `
		DELETE FROM agent_models 
		WHERE agent_id NOT IN (
			SELECT agent_id FROM (
				SELECT DISTINCT ON (hostname) agent_id 
				FROM agent_models 
				ORDER BY hostname, last_auth_at DESC, registered_at DESC
			) tmp
		)`
	return r.db.WithContext(ctx).Exec(query).Error
}

func toAgentEntity(m *AgentModel) *entity.AgentRegistry {
	var metadata map[string]string
	json.Unmarshal([]byte(m.Metadata), &metadata)

	return &entity.AgentRegistry{
		AgentID:      m.AgentID,
		Hostname:     m.Hostname,
		IPAddress:    m.IPAddress,
		AgentVersion: m.AgentVersion,
		AccessToken:  m.AccessToken,
		TokenExpiry:  m.TokenExpiry,
		Status:       entity.AgentStatus(m.Status),
		RegisteredAt: m.RegisteredAt,
		LastAuthAt:   m.LastAuthAt,
		Metadata:     metadata,
	}
}

// Host Repository
type PostgresHostRepository struct {
	db *gorm.DB
}

func NewPostgresHostRepository(db *gorm.DB) repository.HostRepository {
	return &PostgresHostRepository{db: db}
}

func (r *PostgresHostRepository) Create(ctx context.Context, host *entity.Host) error {
	metadata, _ := json.Marshal(host.Metadata)
	model := HostModel{
		ID:           host.ID,
		Hostname:     host.Hostname,
		IPAddress:    host.IPAddress,
		AgentID:      host.AgentID,
		AgentVersion: host.AgentVersion,
		Status:       string(host.Status),
		Metadata:     string(metadata),
		CreatedAt:    host.CreatedAt,
		UpdatedAt:    host.UpdatedAt,
		LastSeenAt:   host.LastSeenAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *PostgresHostRepository) Get(ctx context.Context, hostname string) (*entity.Host, error) {
	var m HostModel
	if err := r.db.WithContext(ctx).First(&m, "hostname = ?", hostname).Error; err != nil {
		return nil, err
	}
	return toHostEntity(&m), nil
}

func (r *PostgresHostRepository) GetAll(ctx context.Context) ([]*entity.Host, error) {
	var models []HostModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	res := make([]*entity.Host, len(models))
	for i, m := range models {
		res[i] = toHostEntity(&m)
	}
	return res, nil
}

func (r *PostgresHostRepository) Update(ctx context.Context, host *entity.Host) error {
	metadata, _ := json.Marshal(host.Metadata)
	updates := map[string]interface{}{
		"ip_address":    host.IPAddress,
		"agent_version": host.AgentVersion,
		"status":        string(host.Status),
		"metadata":      string(metadata),
		"updated_at":    host.UpdatedAt,
		"last_seen_at":  host.LastSeenAt,
	}
	return r.db.WithContext(ctx).Model(&HostModel{}).Where("hostname = ?", host.Hostname).Updates(updates).Error
}

func (r *PostgresHostRepository) Delete(ctx context.Context, hostname string) error {
	return r.db.WithContext(ctx).Delete(&HostModel{}, "hostname = ?", hostname).Error
}

func toHostEntity(m *HostModel) *entity.Host {
	var metadata map[string]string
	json.Unmarshal([]byte(m.Metadata), &metadata)
	return &entity.Host{
		ID:           m.ID,
		Hostname:     m.Hostname,
		IPAddress:    m.IPAddress,
		AgentID:      m.AgentID,
		AgentVersion: m.AgentVersion,
		Status:       entity.HostStatus(m.Status),
		Metadata:     metadata,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		LastSeenAt:   m.LastSeenAt,
	}
}

// User Repository
type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) repository.UserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	model := UserModel{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *PostgresUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) error {
	updates := map[string]interface{}{
		"username":      user.Username,
		"email":         user.Email,
		"role":          user.Role,
		"password_hash": user.PasswordHash,
		"updated_at":    user.UpdatedAt,
	}
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", user.ID).Updates(updates).Error
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	var m UserModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *PostgresUserRepository) List(ctx context.Context) ([]*entity.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*entity.User, len(models))
	for i, m := range models {
		result[i] = toUserEntity(&m)
	}
	return result, nil
}

func toUserEntity(m *UserModel) *entity.User {
	return &entity.User{
		ID:           m.ID,
		Email:        m.Email,
		Username:     m.Username,
		Role:         m.Role,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// Policy Repository
type PostgresPolicyRepository struct {
	db *gorm.DB
}

func NewPostgresPolicyRepository(db *gorm.DB) repository.PolicyRepository {
	return &PostgresPolicyRepository{db: db}
}

func (r *PostgresPolicyRepository) Create(policy *entity.Policy) error {
	thresholds, _ := json.Marshal(policy.Thresholds)
	actions, _ := json.Marshal(policy.Actions)
	metadata, _ := json.Marshal(policy.Metadata)
	model := PolicyModel{
		PolicyID:    policy.PolicyID,
		Name:        policy.Name,
		Description: policy.Description,
		Enabled:     policy.Enabled,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt,
		Thresholds:  string(thresholds),
		Actions:     string(actions),
		Metadata:    string(metadata),
	}
	return r.db.Create(&model).Error
}

func (r *PostgresPolicyRepository) Update(policy *entity.Policy) error {
	thresholds, _ := json.Marshal(policy.Thresholds)
	actions, _ := json.Marshal(policy.Actions)
	metadata, _ := json.Marshal(policy.Metadata)
	updates := map[string]interface{}{
		"name":        policy.Name,
		"description": policy.Description,
		"enabled":     policy.Enabled,
		"updated_at":  policy.UpdatedAt,
		"thresholds":  string(thresholds),
		"actions":     string(actions),
		"metadata":    string(metadata),
	}
	return r.db.Model(&PolicyModel{}).Where("policy_id = ?", policy.PolicyID).Updates(updates).Error
}

func (r *PostgresPolicyRepository) Delete(policyID string) error {
	return r.db.Delete(&PolicyModel{}, "policy_id = ?", policyID).Error
}

func (r *PostgresPolicyRepository) GetByID(policyID string) (*entity.Policy, error) {
	var m PolicyModel
	if err := r.db.First(&m, "policy_id = ?", policyID).Error; err != nil {
		return nil, err
	}
	return toPolicyEntity(&m), nil
}

func (r *PostgresPolicyRepository) GetAll(page, pageSize int) ([]*entity.Policy, int, error) {
	var models []PolicyModel
	var total int64
	r.db.Model(&PolicyModel{}).Count(&total)
	offset := (page - 1) * pageSize
	if err := r.db.Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	res := make([]*entity.Policy, len(models))
	for i, m := range models {
		res[i] = toPolicyEntity(&m)
	}
	return res, int(total), nil
}

func (r *PostgresPolicyRepository) GetByAgent(agentID string) ([]*entity.Policy, error) {
	// For simplicity, assuming a simple many-to-many relationship or similar
	// In a real implementation, we might need an AgentPolicies table
	// For now, return empty or implement a simple table if needed
	return nil, nil
}

func (r *PostgresPolicyRepository) ApplyToAgent(policyID, agentID string) error {
	return nil // TODO: Implement AgentPolicies table
}

func (r *PostgresPolicyRepository) UnapplyFromAgent(policyID, agentID string) error {
	return nil // TODO: Implement AgentPolicies table
}

func toPolicyEntity(m *PolicyModel) *entity.Policy {
	var thresholds map[string]string
	var actions []string
	var metadata map[string]string
	json.Unmarshal([]byte(m.Thresholds), &thresholds)
	json.Unmarshal([]byte(m.Actions), &actions)
	json.Unmarshal([]byte(m.Metadata), &metadata)

	return &entity.Policy{
		PolicyID:    m.PolicyID,
		Name:        m.Name,
		Description: m.Description,
		Enabled:     m.Enabled,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Thresholds:  thresholds,
		Actions:     actions,
		Metadata:    metadata,
	}
}
