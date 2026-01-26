// Package repository defines repository interfaces
package repository

import (
	"context"
	"smart-monitor/backend/internal/domain/entity"
)

// UserRepository defines persistence for users
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	List(ctx context.Context) ([]*entity.User, error)
}
