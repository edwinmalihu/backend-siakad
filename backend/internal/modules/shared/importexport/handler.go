package importexport

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/response"
)

type Handler struct {
	db       *sql.DB
	auditLog *auditlogs.Repository
}

func NewHandler(db *sql.DB, auditLog *auditlogs.Repository) *Handler {
	return &Handler{db: db, auditLog: auditLog}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/shared/import/{module}/template", h.DownloadTemplate)
	mux.HandleFunc("POST /api/v1/shared/import/{module}", h.Import)
	mux.HandleFunc("GET /api/v1/shared/export/{module}", h.Export)
}

func (h *Handler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	module := strings.TrimSpace(r.PathValue("module"))

	var (
		data []byte
		err  error
	)

	switch module {
	case "students":
		data, err = generateStudentsTemplate()
	case "teachers":
		data, err = generateTeachersTemplate()
	case "departments":
		data, err = generateDepartmentsTemplate()
	case "grade-levels":
		data, err = generateGradeLevelsTemplate()
	case "academic-years":
		data, err = generateAcademicYearsTemplate()
	default:
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("module '%s' does not support import", module))
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate template: %s", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=template_import_%s.xlsx", module))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	module := strings.TrimSpace(r.PathValue("module"))

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "failed to parse uploaded file")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "failed to read uploaded file")
		return
	}

	var result *ImportResult

	switch module {
	case "students":
		result, err = importStudents(h.db, fileBytes)
	case "teachers":
		result, err = importTeachers(h.db, fileBytes)
	case "departments":
		result, err = importDepartments(h.db, fileBytes)
	case "grade-levels":
		result, err = importGradeLevels(h.db, fileBytes)
	case "academic-years":
		result, err = importAcademicYears(h.db, fileBytes)
	default:
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("module '%s' does not support import", module))
		return
	}

	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAudit(r.Context(), r, "import", module, map[string]any{
		"total_rows":    result.TotalRows,
		"skipped_rows":  result.SkippedRows,
		"success_count": result.SuccessCount,
		"error_count":   result.ErrorCount,
	})

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    result,
	})
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	module := strings.TrimSpace(r.PathValue("module"))
	filters := parseExportFilters(r.URL.Query())

	var (
		data []byte
		err  error
	)

	switch module {
	case "students":
		data, err = exportStudents(h.db, filters)
	case "teachers":
		data, err = exportTeachers(h.db, filters)
	case "departments":
		data, err = exportDepartments(h.db, filters)
	case "grade-levels":
		data, err = exportGradeLevels(h.db, filters)
	case "academic-years":
		data, err = exportAcademicYears(h.db, filters)
	default:
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("module '%s' does not support export", module))
		return
	}

	if err != nil {
		response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to export data: %s", err.Error()))
		return
	}

	h.logAudit(r.Context(), r, "export", module, map[string]any{
		"filters": filters,
	})

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export_%s.xlsx", module))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) logAudit(ctx context.Context, r *http.Request, action, entityType string, payload map[string]any) {
	if h.auditLog == nil {
		return
	}

	payloadJSON := ""
	if payload != nil {
		if bytes, err := json.Marshal(payload); err == nil {
			payloadJSON = string(bytes)
		}
	}

	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ipAddress = strings.Split(forwarded, ",")[0]
	} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		ipAddress = realIP
	}

	log := &auditlogs.AuditLog{
		Module:      "importexport",
		Action:      action,
		EntityType:  entityType,
		PayloadJSON: payloadJSON,
		IPAddress:   ipAddress,
	}

	_ = h.auditLog.Create(ctx, log)
}
