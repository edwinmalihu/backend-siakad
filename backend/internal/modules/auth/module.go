package auth

import (
	"database/sql"
	"net/http"
	"time"

	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/response"
)

type Module struct {
	db      *sql.DB
	handler *Handler
}

func NewModule(db *sql.DB, tokenSecret string, tokenTTL time.Duration) Module {
	module := Module{
		db: db,
	}

	if db != nil {
		repo := NewRepository(db)
		service := NewService(tokenSecret, tokenTTL)
		auditLogRepo := auditlogs.NewRepository(db)
		revokedRepo := NewRevokedTokenRepository(db)
		module.handler = NewHandler(repo, service, auditLogRepo, revokedRepo)
	}

	return module
}

func (Module) Name() string {
	return "auth"
}

func (Module) BasePath() string {
	return "/api/v1/auth"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/auth/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "auth module is ready",
			"db":      m.db != nil,
		})
	})

	if m.handler != nil {
		m.handler.RegisterRoutes(mux)
		return
	}

	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
