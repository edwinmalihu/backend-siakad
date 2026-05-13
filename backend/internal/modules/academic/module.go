package academic

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/academic/homeroomassignments"
	"siakad/backend/internal/modules/academic/schedules"
	"siakad/backend/internal/modules/academic/subjects"
	"siakad/backend/internal/modules/academic/teachers"
	"siakad/backend/internal/response"
)

type Module struct {
	homeroomAssignmentHandler *homeroomassignments.Handler
	db                        *sql.DB
	scheduleHandler           *schedules.Handler
	subjectHandler            *subjects.Handler
	teacherHandler            *teachers.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{
		db: db,
	}

	if db != nil {
		module.homeroomAssignmentHandler = homeroomassignments.NewHandler(db)
		module.scheduleHandler = schedules.NewHandler(db)
		module.subjectHandler = subjects.NewHandler(db)
		module.teacherHandler = teachers.NewHandler(db)
	}

	return module
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
			"db":      m.db != nil,
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

	if m.homeroomAssignmentHandler != nil {
		m.homeroomAssignmentHandler.RegisterRoutes(mux)
	}
	if m.teacherHandler != nil {
		m.teacherHandler.RegisterRoutes(mux)
	}
	if m.scheduleHandler != nil {
		m.scheduleHandler.RegisterRoutes(mux)
	}
	if m.subjectHandler != nil {
		m.subjectHandler.RegisterRoutes(mux)
	}

	if m.homeroomAssignmentHandler != nil || m.teacherHandler != nil || m.scheduleHandler != nil || m.subjectHandler != nil {
		return
	}

	mux.HandleFunc("GET /api/v1/academic/homeroom-assignments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/homeroom-assignments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/homeroom-assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/homeroom-assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/homeroom-assignments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/teachers", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/teachers", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/teachers/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/teachers/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/teachers/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/schedules", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/schedules", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/subjects", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/subjects", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/subjects/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/subjects/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/subjects/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
