// Package http provides administrative HTTP handlers
package http

import (
	"encoding/json"
	"net/http"

	"smart-monitor/backend/internal/domain/service"
)

// AdminUserHandler supports creating users by admins with custom roles
// Route: POST /tools/users
// Body: {"email":"...","username":"...","password":"...","role":"admin|operator|viewer"}
type AdminUserHandler struct {
	userAuth *service.UserAuthService
}

func NewAdminUserHandler(userAuth *service.UserAuthService) *AdminUserHandler {
	return &AdminUserHandler{userAuth: userAuth}
}

func (h *AdminUserHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"success": false, "message": "invalid json body"})
		return
	}
	user, err := h.userAuth.SignUp(r.Context(), req.Email, req.Username, req.Password, req.Role)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"success": false, "message": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"user": map[string]any{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"role":       user.Role,
			"created_at": user.CreatedAt.Unix(),
		},
	})
}
