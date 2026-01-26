// Package persistence implements repository interfaces
package persistence

import (
	"context"
	"fmt"
	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"
	"sync"
)

// InMemoryUserRepository stores users in memory
type InMemoryUserRepository struct {
	mu      sync.RWMutex
	byID    map[string]*entity.User
	byEmail map[string]*entity.User
}

func NewInMemoryUserRepository() repository.UserRepository {
	return &InMemoryUserRepository{byID: make(map[string]*entity.User), byEmail: make(map[string]*entity.User)}
}

func (r *InMemoryUserRepository) Create(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byEmail[user.Email]; exists {
		return fmt.Errorf("user already exists")
	}
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u := r.byEmail[email]
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u := r.byID[id]
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[user.ID] = user
	r.byEmail[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) List(ctx context.Context) ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*entity.User, 0, len(r.byID))
	for _, u := range r.byID {
		out = append(out, u)
	}
	return out, nil
}
