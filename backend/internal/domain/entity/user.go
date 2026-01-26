// Package entity defines core business entities
package entity

import "time"

// User represents a system user for frontend authentication
type User struct {
	ID           string
	Email        string
	Username     string
	Role         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id, email, username, role, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           id,
		Email:        email,
		Username:     username,
		Role:         role,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) Touch() { u.UpdatedAt = time.Now() }
