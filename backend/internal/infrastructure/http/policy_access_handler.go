// Package http provides HTTP handlers for policy access management
package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"smart-monitor/backend/internal/domain/service"
)

// PolicyAccessHandler manages allowed users for policies
type PolicyAccessHandler struct {
	policyService *service.PolicyService
	userAuth      *service.UserAuthService
}

func NewPolicyAccessHandler(policyService *service.PolicyService, userAuth *service.UserAuthService) *PolicyAccessHandler {
	return &PolicyAccessHandler{policyService: policyService, userAuth: userAuth}
}

// ServeHTTP handles routes:
// POST /v1/policies/{policy_id}/allowed-users/add {"user_id":"..."}
// POST /v1/policies/{policy_id}/allowed-users/remove {"user_id":"..."}
// GET  /v1/policies/{policy_id}/allowed-users
func (h *PolicyAccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	// Expect path starting with /v1/policies/
	if !strings.HasPrefix(path, "/v1/policies/") {
		http.NotFound(w, r)
		return
	}
	rest := strings.TrimPrefix(path, "/v1/policies/")
	// rest should be like {policy_id}/allowed-users[/action]
	parts := strings.Split(rest, "/")
	if len(parts) < 2 || parts[1] != "allowed-users" {
		http.NotFound(w, r)
		return
	}
	policyID := parts[0]

	// GET list
	if r.Method == http.MethodGet && len(parts) == 2 {
		ids, err := h.policyService.ListAllowedUsers(policyID)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{"success": false, "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "policy_id": policyID, "allowed_user_ids": ids})
		return
	}

	// POST add/remove
	if r.Method == http.MethodPost && len(parts) == 3 {
		var body struct {
			UserID string `json:"user_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{"success": false, "message": "Invalid request body"})
			return
		}
		action := parts[2]
		var err error
		switch action {
		case "add":
			err = h.policyService.AddAllowedUserToPolicy(policyID, body.UserID)
		case "remove":
			err = h.policyService.RemoveAllowedUserFromPolicy(policyID, body.UserID)
		default:
			http.NotFound(w, r)
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{"success": false, "message": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "policy_id": policyID, "user_id": body.UserID, "action": action})
		return
	}

	http.NotFound(w, r)
}
