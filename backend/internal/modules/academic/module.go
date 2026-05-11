package academic

import (
	"net/http"

	"siakad/backend/internal/response"
)

type Module struct{}

func NewModule() Module {
	return Module{}
}

func (Module) Name() string {
	return "academic"
}

func (Module) BasePath() string {
	return "/api/v1/academic"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/academic/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "academic module is ready",
		})
	})

	mux.HandleFunc("GET /api/v1/academic/resources", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"resources": []string{
				"teachers",
				"subjects",
				"homeroom_assignments",
				"schedules",
				"assessment_components",
				"student_assessments",
				"student_grades",
			},
		})
	})
}
