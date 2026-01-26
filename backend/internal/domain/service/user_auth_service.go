// Package service defines domain services
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"smart-monitor/backend/internal/domain/entity"
	"smart-monitor/backend/internal/domain/repository"

	"github.com/golang-jwt/jwt/v5"
)

// UserAuthService handles user signup/signin and token generation
type UserAuthService struct {
	users     repository.UserRepository
	jwtSecret []byte
}

func NewUserAuthService(users repository.UserRepository, jwtSecret string) *UserAuthService {
	return &UserAuthService{users: users, jwtSecret: []byte(jwtSecret)}
}

// SignUp registers a new user
func (s *UserAuthService) SignUp(ctx context.Context, email, username, password, role string) (*entity.User, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password required")
	}
	if role == "" {
		role = "viewer"
	}
	if role != "admin" && role != "operator" && role != "viewer" {
		return nil, fmt.Errorf("invalid role")
	}
	id := generateUserID(email)
	hash := hashPassword(password)
	user := entity.NewUser(id, email, username, role, hash)
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// SignIn validates credentials and returns a JWT token
func (s *UserAuthService) SignIn(ctx context.Context, email, password string) (string, *entity.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, fmt.Errorf("invalid credentials")
	}
	if !checkPassword(password, u.PasswordHash) {
		return "", nil, fmt.Errorf("invalid credentials")
	}

	claims := jwt.MapClaims{
		"sub":      u.ID,
		"email":    u.Email,
		"username": u.Username,
		"role":     u.Role,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", nil, err
	}
	return signed, u, nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func checkPassword(password, hash string) bool { return hashPassword(password) == hash }

func generateUserID(email string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", email, time.Now().UnixNano())))
	return fmt.Sprintf("user-%x", h[:8])
}

// ParseToken validates JWT and returns claims
func (s *UserAuthService) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	parsed, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
