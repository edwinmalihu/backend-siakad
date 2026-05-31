package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"siakad/backend/internal/response"
)

type contextKey string

const UserContextKey contextKey = "user"

type UserContext struct {
	UserID   uint64
	Username string
	Roles    []string
}

var publicExactPaths = map[string]bool{
	"/health":   true,
	"/api/v1":   true,
	"/openapi.yaml": true,
}

var publicPrefixes = []string{
	"/docs/",
	"/api/v1/auth/",
}

func AuthMiddleware(service *Service, repo *Repository, revokedRepo *RevokedTokenRepository, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if publicExactPaths[path] {
			next.ServeHTTP(w, r)
			return
		}

		for _, prefix := range publicPrefixes {
			if strings.HasPrefix(path, prefix) {
				next.ServeHTTP(w, r)
				return
			}
		}

		token, ok := extractBearerToken(r.Header.Get("Authorization"))
		if !ok {
			response.Error(w, http.StatusUnauthorized, "authorization token is required")
			return
		}

		// Check if token has been revoked (logout)
		if revokedRepo != nil {
			revoked, err := revokedRepo.IsRevoked(r.Context(), token)
			if err == nil && revoked {
				response.Error(w, http.StatusUnauthorized, "token has been revoked")
				return
			}
		}

		claims, err := service.ParseToken(token)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Check if user is still active
		user, err := repo.FindByID(r.Context(), claims.Sub)
		if err != nil {
			if err == sql.ErrNoRows || err == ErrUserNotFound {
				response.Error(w, http.StatusUnauthorized, "user account not found")
				return
			}
			response.Error(w, http.StatusInternalServerError, "failed to verify user")
			return
		}

		if !user.IsActive {
			response.Error(w, http.StatusForbidden, "user account is inactive")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, &UserContext{
			UserID:   claims.Sub,
			Username: claims.Username,
			Roles:    claims.Roles,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) *UserContext {
	user, _ := ctx.Value(UserContextKey).(*UserContext)
	return user
}
