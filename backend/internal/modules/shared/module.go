package shared

import (
	"net/http"

	"siakad/backend/internal/response"
)

type Module struct{}

func NewModule() Module {
	return Module{}
}

func (Module) Name() string {
	return "shared"
}

func (Module) BasePath() string {
	return "/api/v1/shared"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/shared/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "shared module is ready",
		})
	})

	mux.HandleFunc("GET /api/v1/shared/resources", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"resources": []string{
				"announcements",
				"audit_logs",
				"global_search",
				"dashboard",
			},
		})
	})
}
