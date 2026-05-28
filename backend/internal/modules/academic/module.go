package academic

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/academic/assessmentcomponents"
	"siakad/backend/internal/modules/academic/homeroomassignments"
	"siakad/backend/internal/modules/academic/schedules"
	"siakad/backend/internal/modules/academic/studentassessments"
	"siakad/backend/internal/modules/academic/studentgrades"
	"siakad/backend/internal/modules/academic/subjects"
	"siakad/backend/internal/modules/academic/teachers"
	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/response"
)

type Module struct {
	assessmentComponentHandler *assessmentcomponents.Handler
	homeroomAssignmentHandler  *homeroomassignments.Handler
	db                         *sql.DB
	scheduleHandler            *schedules.Handler
	studentAssessmentHandler   *studentassessments.Handler
	studentGradeHandler        *studentgrades.Handler
	subjectHandler             *subjects.Handler
	teacherHandler             *teachers.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{
		db: db,
	}

	if db != nil {
		auditLogRepo := auditlogs.NewRepository(db)
		module.assessmentComponentHandler = assessmentcomponents.NewHandler(db, auditLogRepo)
		module.homeroomAssignmentHandler = homeroomassignments.NewHandler(db, auditLogRepo)
		module.scheduleHandler = schedules.NewHandler(db, auditLogRepo)
		module.studentAssessmentHandler = studentassessments.NewHandler(db, auditLogRepo)
		module.studentGradeHandler = studentgrades.NewHandler(db, auditLogRepo)
		module.subjectHandler = subjects.NewHandler(db, auditLogRepo)
		module.teacherHandler = teachers.NewHandler(db, auditLogRepo)
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
	if m.assessmentComponentHandler != nil {
		m.assessmentComponentHandler.RegisterRoutes(mux)
	}
	if m.studentAssessmentHandler != nil {
		m.studentAssessmentHandler.RegisterRoutes(mux)
	}
	if m.studentGradeHandler != nil {
		m.studentGradeHandler.RegisterRoutes(mux)
	}

	if m.homeroomAssignmentHandler != nil || m.teacherHandler != nil || m.scheduleHandler != nil || m.subjectHandler != nil || m.assessmentComponentHandler != nil || m.studentAssessmentHandler != nil || m.studentGradeHandler != nil {
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
	mux.HandleFunc("GET /api/v1/academic/assessment-components", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/assessment-components", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/assessment-components/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/assessment-components/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/assessment-components/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/student-assessments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/student-assessments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/student-assessments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/student-assessments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/student-assessments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/student-grades", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/academic/student-grades", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/academic/student-grades/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/academic/student-grades/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/academic/student-grades/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
