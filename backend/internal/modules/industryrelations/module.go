package industryrelations

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/industryrelations/alumni"
	"siakad/backend/internal/modules/industryrelations/companies"
	"siakad/backend/internal/modules/industryrelations/industrycategories"
	"siakad/backend/internal/modules/industryrelations/internships"
	"siakad/backend/internal/modules/industryrelations/internshiplogs"
	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/response"
)

type Module struct {
	db                      *sql.DB
	alumniHandler           *alumni.Handler
	companyHandler          *companies.Handler
	industryCategoryHandler *industrycategories.Handler
	internshipHandler       *internships.Handler
	internshipLogHandler    *internshiplogs.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{db: db}
	if db != nil {
		auditLogRepo := auditlogs.NewRepository(db)
		module.alumniHandler = alumni.NewHandler(db, auditLogRepo)
		module.industryCategoryHandler = industrycategories.NewHandler(db, auditLogRepo)
		module.companyHandler = companies.NewHandler(db, auditLogRepo)
		module.internshipHandler = internships.NewHandler(db, auditLogRepo)
		module.internshipLogHandler = internshiplogs.NewHandler(db, auditLogRepo)
	}
	return module
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
			"db":      m.db != nil,
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

	if m.industryCategoryHandler != nil {
		m.industryCategoryHandler.RegisterRoutes(mux)
	}
	if m.companyHandler != nil {
		m.companyHandler.RegisterRoutes(mux)
	}
	if m.internshipHandler != nil {
		m.internshipHandler.RegisterRoutes(mux)
	}
	if m.alumniHandler != nil {
		m.alumniHandler.RegisterRoutes(mux)
	}
	if m.internshipLogHandler != nil {
		m.internshipLogHandler.RegisterRoutes(mux)
	}

	if m.industryCategoryHandler != nil && m.companyHandler != nil && m.internshipHandler != nil && m.alumniHandler != nil && m.internshipLogHandler != nil {
		return
	}

	mux.HandleFunc("GET /api/v1/industry-relations/categories", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/industry-relations/categories", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/industry-relations/categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/industry-relations/categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/industry-relations/categories/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})

	mux.HandleFunc("GET /api/v1/industry-relations/companies", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/industry-relations/companies", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/industry-relations/companies/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/industry-relations/companies/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/industry-relations/companies/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})

	mux.HandleFunc("GET /api/v1/industry-relations/internships", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/industry-relations/internships", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/industry-relations/internships/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/industry-relations/internships/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/industry-relations/internships/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})

	mux.HandleFunc("GET /api/v1/industry-relations/alumni", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/industry-relations/alumni", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/industry-relations/alumni/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/industry-relations/alumni/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/industry-relations/alumni/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/industry-relations/internship-logs", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/industry-relations/internship-logs", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/industry-relations/internship-logs/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/industry-relations/internship-logs/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/industry-relations/internship-logs/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
