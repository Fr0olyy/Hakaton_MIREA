package auth

import (
	"context"
	"net/http"
	"strings"

	"hakaton/backend/internal/models"
)

type contextKey string

const userKey contextKey = "user"

func ContextWithUser(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, userKey, claims)
}

func UserFromContext(ctx context.Context) *Claims {
	claims, _ := ctx.Value(userKey).(*Claims)
	return claims
}

func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, http.StatusUnauthorized, "authorization header is required")
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeAuthError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}
			claims, err := ValidateToken(parts[1], secret)
			if err != nil {
				if err == ErrExpiredToken {
					writeAuthError(w, http.StatusUnauthorized, "token expired")
					return
				}
				writeAuthError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			ctx := ContextWithUser(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

var actionPermissions = map[models.ProjectMemberRole]map[string]bool{
	models.ProjectRoleAdmin: {
		"object:approve": true, "object:exclude": true, "object:relabel": true,
		"object:mark_duplicate": true, "object:send_to_expert": true, "object:add_to_next_train": true,
		"object:expert_approve": true, "object:expert_reject": true,
		"object:comment": true, "object:view": true, "object:review": true,
		"weights:update": true, "benchmark:run": true,
		"report:view": true, "report:export": true,
		"agent:ask": true, "agent:generate_summary": true,
		"manage_members": true, "export": true,
	},
	models.ProjectRoleMLEngineer: {
		"object:view": true, "object:comment": true,
		"analysis:run": true, "metrics:view": true,
		"agent:ask": true, "report:view": true,
	},
	models.ProjectRoleAnnotator: {
		"object:review": true, "object:comment": true, "object:relabel": true,
		"agent:ask": true,
	},
	models.ProjectRoleDomainExpert: {
		"object:send_to_expert": true, "object:expert_approve": true, "object:expert_reject": true,
		"object:comment": true, "object:view": true,
		"agent:ask": true,
	},
	models.ProjectRoleDataAnalyst: {
		"report:view": true, "report:export": true, "metrics:view": true,
		"agent:ask": true,
	},
}

func HasPermission(role models.ProjectMemberRole, action string) bool {
	perms, ok := actionPermissions[role]
	if !ok {
		return false
	}
	return perms[action]
}

func RequirePermission(action string, getRole func(r *http.Request) models.ProjectMemberRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := getRole(r)
			if !HasPermission(role, action) {
				writeAuthError(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
