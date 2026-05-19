package studentaffairs

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/studentaffairs/attendances"
	"siakad/backend/internal/modules/studentaffairs/disciplinecategories"
	"siakad/backend/internal/modules/studentaffairs/disciplinerecords"
	"siakad/backend/internal/modules/studentaffairs/extracurricularmembers"
	"siakad/backend/internal/modules/studentaffairs/extracurriculars"
	"siakad/backend/internal/modules/studentaffairs/studentenrollments"
	"siakad/backend/internal/modules/studentaffairs/studentgraduations"
	"siakad/backend/internal/modules/studentaffairs/studentmutations"
	"siakad/backend/internal/modules/studentaffairs/students"
	"siakad/backend/internal/response"
)

type Module struct {
	db                           *sql.DB
	attendanceHandler            *attendances.Handler
	disciplineCategoryHandler    *disciplinecategories.Handler
	disciplineRecordHandler      *disciplinerecords.Handler
	extracurricularHandler       *extracurriculars.Handler
	extracurricularMemberHandler *extracurricularmembers.Handler
	studentGraduationHandler     *studentgraduations.Handler
	studentHandler               *students.Handler
	studentEnrollmentHandler     *studentenrollments.Handler
	studentMutationHandler       *studentmutations.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{
		db: db,
	}

	if db != nil {
		module.attendanceHandler = attendances.NewHandler(db)
		module.disciplineCategoryHandler = disciplinecategories.NewHandler(db)
		module.disciplineRecordHandler = disciplinerecords.NewHandler(db)
		module.extracurricularHandler = extracurriculars.NewHandler(db)
		module.extracurricularMemberHandler = extracurricularmembers.NewHandler(db)
		module.studentGraduationHandler = studentgraduations.NewHandler(db)
		module.studentHandler = students.NewHandler(db)
		module.studentEnrollmentHandler = studentenrollments.NewHandler(db)
		module.studentMutationHandler = studentmutations.NewHandler(db)
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
				"discipline_categories",
				"discipline_records",
				"extracurriculars",
				"extracurricular_members",
			},
		})
	})

	if m.studentHandler != nil {
		m.studentHandler.RegisterRoutes(mux)
	}
	if m.studentEnrollmentHandler != nil {
		m.studentEnrollmentHandler.RegisterRoutes(mux)
	}
	if m.studentMutationHandler != nil {
		m.studentMutationHandler.RegisterRoutes(mux)
	}
	if m.studentGraduationHandler != nil {
		m.studentGraduationHandler.RegisterRoutes(mux)
	}
	if m.attendanceHandler != nil {
		m.attendanceHandler.RegisterRoutes(mux)
	}
	if m.disciplineCategoryHandler != nil {
		m.disciplineCategoryHandler.RegisterRoutes(mux)
	}
	if m.disciplineRecordHandler != nil {
		m.disciplineRecordHandler.RegisterRoutes(mux)
	}
	if m.extracurricularHandler != nil {
		m.extracurricularHandler.RegisterRoutes(mux)
	}
	if m.extracurricularMemberHandler != nil {
		m.extracurricularMemberHandler.RegisterRoutes(mux)
	}

	if m.studentHandler != nil &&
		m.studentEnrollmentHandler != nil &&
		m.studentMutationHandler != nil &&
		m.studentGraduationHandler != nil &&
		m.attendanceHandler != nil &&
		m.disciplineCategoryHandler != nil &&
		m.disciplineRecordHandler != nil &&
		m.extracurricularHandler != nil &&
		m.extracurricularMemberHandler != nil {
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
	mux.HandleFunc("GET /api/v1/student-affairs/enrollments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/enrollments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/enrollments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/enrollments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/enrollments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/mutations", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/mutations", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/mutations/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/mutations/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/mutations/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/graduations", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/graduations", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/graduations/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/graduations/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/graduations/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/attendances", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/attendances", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/attendances/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/attendances/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/attendances/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/discipline-categories", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/discipline-categories", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/discipline-categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/discipline-categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/discipline-categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/discipline-records", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/discipline-records", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/discipline-records/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/discipline-records/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/discipline-records/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/extracurriculars", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/extracurriculars", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/extracurriculars/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/extracurriculars/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/extracurriculars/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/extracurricular-members", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/student-affairs/extracurricular-members", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/student-affairs/extracurricular-members/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/student-affairs/extracurricular-members/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/student-affairs/extracurricular-members/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
