package auth

import (
	"net/http"

	"siakad/backend/internal/response"
)

type Module struct{}

func NewModule() Module {
	return Module{}
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
		})
	})

	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusNotImplemented, map[string]any{
			"success": false,
			"module":  m.Name(),
			"message": "login handler is not implemented yet",
		})
	})
}
