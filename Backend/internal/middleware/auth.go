package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/auction-studio/backend/internal/auth"
	"github.com/auction-studio/backend/internal/models"
)

type contextKey string

const (
	UserContextKey  contextKey = "user"
	RoleContextKey  contextKey = "role"
	EmailContextKey contextKey = "email"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := auth.ValidateToken(parts[1], secret)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleContextKey, claims.Role)
			ctx = context.WithValue(ctx, EmailContextKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(UserContextKey)
		if userID == nil || userID == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireRole(roles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleContextKey).(models.UserRole)
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}

func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserContextKey).(string); ok {
		return val
	}
	return ""
}

func GetUserRole(ctx context.Context) models.UserRole {
	if val, ok := ctx.Value(RoleContextKey).(models.UserRole); ok {
		return val
	}
	return ""
}
