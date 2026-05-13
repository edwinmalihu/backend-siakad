package master

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/master/academicyears"
	"siakad/backend/internal/modules/master/classes"
	"siakad/backend/internal/modules/master/departments"
	"siakad/backend/internal/modules/master/gradelevels"
	"siakad/backend/internal/modules/master/semesters"
	"siakad/backend/internal/response"
)

type Module struct {
	db                  *sql.DB
	academicYearHandler *academicyears.Handler
	classHandler        *classes.Handler
	departmentHandler   *departments.Handler
	gradeLevelHandler   *gradelevels.Handler
	semesterHandler     *semesters.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{
		db: db,
	}

	if db != nil {
		module.academicYearHandler = academicyears.NewHandler(db)
		module.classHandler = classes.NewHandler(db)
		module.departmentHandler = departments.NewHandler(db)
		module.gradeLevelHandler = gradelevels.NewHandler(db)
		module.semesterHandler = semesters.NewHandler(db)
	}

	return module
}

func (Module) Name() string {
	return "master"
}

func (Module) BasePath() string {
	return "/api/v1/master"
}

func (m Module) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/master/health", func(w http.ResponseWriter, r *http.Request) {
		dbReady := m.db != nil
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"message": "master module is ready",
			"db":      dbReady,
		})
	})

	mux.HandleFunc("GET /api/v1/master/reference-data", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"module":  m.Name(),
			"resources": []string{
				"academic_years",
				"semesters",
				"departments",
				"grade_levels",
				"classes",
				"rooms",
			},
		})
	})

	if m.academicYearHandler != nil {
		m.academicYearHandler.RegisterRoutes(mux)
	}
	if m.classHandler != nil {
		m.classHandler.RegisterRoutes(mux)
	}
	if m.departmentHandler != nil {
		m.departmentHandler.RegisterRoutes(mux)
	}
	if m.gradeLevelHandler != nil {
		m.gradeLevelHandler.RegisterRoutes(mux)
	}
	if m.semesterHandler != nil {
		m.semesterHandler.RegisterRoutes(mux)
	}

	if m.academicYearHandler != nil || m.classHandler != nil || m.departmentHandler != nil || m.gradeLevelHandler != nil || m.semesterHandler != nil {
		return
	}

	mux.HandleFunc("GET /api/v1/master/academic-years", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/master/academic-years", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/academic-years/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/master/academic-years/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/master/academic-years/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/classes", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/master/classes", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/classes/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/master/classes/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/master/classes/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/semesters", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/master/semesters", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/semesters/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/master/semesters/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/master/semesters/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/departments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/master/departments", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/master/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/master/departments/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/grade-levels", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/master/grade-levels", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/master/grade-levels/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/master/grade-levels/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/master/grade-levels/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
