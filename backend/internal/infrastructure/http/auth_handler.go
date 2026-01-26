// Package http provides HTTP handlers
package http

import (
	"encoding/json"
	"net/http"

	"smart-monitor/backend/internal/domain/service"
)

// AuthHandler provides sign-in and sign-up endpoints
type AuthHandler struct {
	users *service.UserAuthService
}

func NewAuthHandler(users *service.UserAuthService) *AuthHandler { return &AuthHandler{users: users} }

// SignUp: POST /auth/signup { email, username, password }
func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	user, err := h.users.SignUp(r.Context(), body.Email, body.Username, body.Password, body.Role)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"user": map[string]interface{}{"id": user.ID, "email": user.Email, "username": user.Username, "role": user.Role}})
}

// SignIn: POST /auth/signin { email, password }
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	token, user, err := h.users.SignIn(r.Context(), body.Email, body.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"token": token, "user": map[string]interface{}{"id": user.ID, "email": user.Email, "username": user.Username, "role": user.Role}})
}
