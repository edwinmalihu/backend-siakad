package disciplinerecords

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"siakad/backend/internal/modules/auth"
	"siakad/backend/internal/modules/shared/auditlogs"
	"siakad/backend/internal/response"
)

type Handler struct {
	repo     *Repository
	auditLog *auditlogs.Repository
}

func NewHandler(db *sql.DB, auditLog *auditlogs.Repository) *Handler {
	return &Handler{
		repo:     NewRepository(db),
		auditLog: auditLog,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/student-affairs/discipline-records", h.List)
	mux.HandleFunc("POST /api/v1/student-affairs/discipline-records", h.Create)
	mux.HandleFunc("GET /api/v1/student-affairs/discipline-records/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/student-affairs/discipline-records/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/student-affairs/discipline-records/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	incidentDate := strings.TrimSpace(r.URL.Query().Get("incident_date"))
	if incidentDate != "" {
		if _, err := time.Parse("2006-01-02", incidentDate); err != nil {
			response.Error(w, http.StatusBadRequest, "incident_date must use YYYY-MM-DD format")
			return
		}
	}
	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	categoryID, ok := parseOptionalUint64(w, r.URL.Query().Get("discipline_category_id"), "discipline_category_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, incidentDate, studentID, categoryID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
		"meta": map[string]any{
			"total": len(items),
		},
	})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "discipline record not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    item,
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}
	item, err := buildDisciplineRecord(req.StudentID, req.DisciplineCategoryID, req.RecordedBy, req.IncidentDate, req.Description, req.ActionTaken)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var userID *uint64
	if user != nil {
		uid := user.UserID
		userID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "create", "discipline_record", created.ID, userID, req)

	response.JSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    created,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}
	item, err := buildDisciplineRecord(req.StudentID, req.DisciplineCategoryID, req.RecordedBy, req.IncidentDate, req.Description, req.ActionTaken)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var userID *uint64
	if user != nil {
		uid := user.UserID
		userID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "update", "discipline_record", id, userID, req)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    updated,
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "discipline record not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var userID *uint64
	if user != nil {
		uid := user.UserID
		userID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "delete", "discipline_record", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "discipline record deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "discipline record not found")
	case errors.Is(err, ErrStudentNotFound):
		response.Error(w, http.StatusBadRequest, "student not found")
	case errors.Is(err, ErrDisciplineCategoryNotFound):
		response.Error(w, http.StatusBadRequest, "discipline category not found")
	case errors.Is(err, ErrRecordedByNotFound):
		response.Error(w, http.StatusBadRequest, "recorded_by user not found")
	default:
		response.Error(w, http.StatusInternalServerError, err.Error())
	}
}

func parseID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid resource id")
		return 0, false
	}
	return id, true
}

func parseOptionalUint64(w http.ResponseWriter, raw, field string) (uint64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}
	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", field))
		return 0, false
	}
	return parsed, true
}

func buildDisciplineRecord(studentID, categoryID, recordedBy uint64, incidentDate, description, actionTaken string) (DisciplineRecord, error) {
	if studentID == 0 {
		return DisciplineRecord{}, fmt.Errorf("student_id is required")
	}
	if categoryID == 0 {
		return DisciplineRecord{}, fmt.Errorf("discipline_category_id is required")
	}
	if strings.TrimSpace(incidentDate) == "" {
		return DisciplineRecord{}, fmt.Errorf("incident_date is required")
	}
	parsedIncidentDate, err := time.Parse("2006-01-02", strings.TrimSpace(incidentDate))
	if err != nil {
		return DisciplineRecord{}, fmt.Errorf("incident_date must use YYYY-MM-DD format")
	}

	var recordedByValue *uint64
	if recordedBy > 0 {
		recordedByValue = &recordedBy
	}

	return DisciplineRecord{
		StudentID:            studentID,
		DisciplineCategoryID: categoryID,
		RecordedBy:           recordedByValue,
		IncidentDate:         parsedIncidentDate,
		Description:          strings.TrimSpace(description),
		ActionTaken:          strings.TrimSpace(actionTaken),
	}, nil
}
