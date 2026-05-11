package studentaffairs

import (
	"net/http"

	"siakad/backend/internal/response"
)

type Module struct{}

func NewModule() Module {
	return Module{}
}

func (Module) Name() string {
	return "student_affairs"
}

func (Module) BasePath() string {
	return "/api/v1/student-affairs"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/student-affairs/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "student affairs module is ready",
		})
	})

	mux.HandleFunc("GET /api/v1/student-affairs/resources", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"resources": []string{
				"students",
				"student_enrollments",
				"student_mutations",
				"student_graduations",
				"attendances",
				"discipline_records",
				"extracurriculars",
			},
		})
	})
}
