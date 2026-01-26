// Package http provides HTTP handlers and middleware
package http

import (
	"net/http"
	"strings"

	"smart-monitor/backend/internal/domain/service"
)

// RequireRoles wraps a handler and enforces that the request has a valid Bearer token with one of the allowed roles.
func RequireRoles(auth *service.UserAuthService, allowed []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authz, "Bearer ")
		claims, err := auth.ParseToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		roleVal, ok := claims["role"].(string)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		for _, a := range allowed {
			if roleVal == a {
				next(w, r)
				return
			}
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}
