package importexport

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"

	"siakad/backend/internal/response"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/shared/import/{module}/template", h.DownloadTemplate)
	mux.HandleFunc("POST /api/v1/shared/import/{module}", h.Import)
	mux.HandleFunc("GET /api/v1/shared/export/{module}", h.Export)
}

func (h *Handler) DownloadTemplate(w http.ResponseWriter, r *http.Request) {
	module := strings.TrimSpace(r.PathValue("module"))

	switch module {
	case "students":
		data, err := generateStudentsTemplate()
		if err != nil {
			response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate template: %s", err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=template_import_students.xlsx")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	default:
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("module '%s' does not support import", module))
	}
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

	switch module {
	case "students":
		result, err := importStudents(h.db, fileBytes)
		if err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    result,
		})
	default:
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("module '%s' does not support import", module))
	}
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	module := strings.TrimSpace(r.PathValue("module"))

	switch module {
	case "students":
		data, err := exportStudents(h.db)
		if err != nil {
			response.Error(w, http.StatusInternalServerError, fmt.Sprintf("failed to export data: %s", err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=export_students.xlsx")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	default:
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("module '%s' does not support export", module))
	}
}
