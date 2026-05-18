package studentaffairs

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/studentaffairs/students"
	"siakad/backend/internal/response"
)

type Module struct {
	db             *sql.DB
	studentHandler *students.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{
		db: db,
	}

	if db != nil {
		module.studentHandler = students.NewHandler(db)
	}

	return module
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
			"db":      m.db != nil,
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

	if m.studentHandler != nil {
		m.studentHandler.RegisterRoutes(mux)
		return
	}

	mux.HandleFunc("GET /api/v1/student-affairs/students", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/students", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/students/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
