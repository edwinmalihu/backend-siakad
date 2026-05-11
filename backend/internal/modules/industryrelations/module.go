package industryrelations

import (
	"net/http"

	"siakad/backend/internal/response"
)

type Module struct{}

func NewModule() Module {
	return Module{}
}

func (Module) Name() string {
	return "industry_relations"
}

func (Module) BasePath() string {
	return "/api/v1/industry-relations"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/industry-relations/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "industry relations module is ready",
		})
	})

	mux.HandleFunc("GET /api/v1/industry-relations/resources", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"resources": []string{
				"industry_categories",
				"companies",
				"internships",
				"internship_logs",
				"alumni",
			},
		})
	})
}
