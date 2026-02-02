package postgres

import (
	"time"
)

// AgentModel is the GORM model for AgentRegistry
type AgentModel struct {
	AgentID      string `gorm:"primaryKey"`
	Hostname     string `gorm:"uniqueIndex"`
	IPAddress    string
	AgentVersion string
	AccessToken  string `gorm:"uniqueIndex"`
	TokenExpiry  time.Time
	Status       string
	Blocked      bool
	BlockReason  string
	Metadata     string // JSON encoded
	RegisteredAt time.Time
	LastAuthAt   time.Time
}

// HostModel is the GORM model for Host
type HostModel struct {
	ID           string `gorm:"primaryKey"`
	Hostname     string `gorm:"uniqueIndex"`
	IPAddress    string
	AgentID      string
	AgentVersion string
	Status       string
	Metadata     string // JSON encoded
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastSeenAt   time.Time
}

// PolicyModel is the GORM model for Policy
type PolicyModel struct {
	PolicyID    string `gorm:"primaryKey"`
	Name        string
	Description string
	Thresholds  string // JSON encoded
	Actions     string // JSON encoded
	Metadata    string // JSON encoded
	Enabled     bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserModel is the GORM model for User
type UserModel struct {
	ID           string `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex"`
	Username     string `gorm:"uniqueIndex"`
	Role         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
