package shared

import (
	"database/sql"
	"net/http"

	"siakad/backend/internal/modules/shared/announcements"
	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/modules/shared/importexport"
	"siakad/backend/internal/modules/shared/studentsearch"
	"siakad/backend/internal/response"
)

type Module struct {
	db                      *sql.DB
	announcementHandler     *announcements.Handler
	auditLogHandler         *auditlogs.Handler
	searchHandler           *studentsearch.Handler
	importExportHandler     *importexport.Handler
}

func NewModule(db *sql.DB) Module {
	module := Module{db: db}
	if db != nil {
		auditLogRepo := auditlogs.NewRepository(db)
		module.announcementHandler = announcements.NewHandler(db)
		module.auditLogHandler = auditlogs.NewHandler(db)
		module.searchHandler = studentsearch.NewHandler(db)
		module.importExportHandler = importexport.NewHandler(db, auditLogRepo)
	}
	return module
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

	if m.announcementHandler != nil {
		m.announcementHandler.RegisterRoutes(mux)
	}
	if m.auditLogHandler != nil {
		m.auditLogHandler.RegisterRoutes(mux)
	}
	if m.searchHandler != nil {
		m.searchHandler.RegisterRoutes(mux)
	}
	if m.importExportHandler != nil {
		m.importExportHandler.RegisterRoutes(mux)
	}

	if m.announcementHandler != nil && m.auditLogHandler != nil && m.searchHandler != nil && m.importExportHandler != nil {
		return
	}

	mux.HandleFunc("GET /api/v1/shared/announcements", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/shared/announcements", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/announcements/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("PUT /api/v1/shared/announcements/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("DELETE /api/v1/shared/announcements/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/student-search", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/student-search/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/audit-logs", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/audit-logs/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/import/{module}/template", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("POST /api/v1/shared/import/{module}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
	mux.HandleFunc("GET /api/v1/shared/export/{module}", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusServiceUnavailable, "database connection is not configured")
	})
}
