package studentgraduations

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
	mux.HandleFunc("GET /api/v1/student-affairs/graduations", h.List)
	mux.HandleFunc("POST /api/v1/student-affairs/graduations", h.Create)
	mux.HandleFunc("GET /api/v1/student-affairs/graduations/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/student-affairs/graduations/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/student-affairs/graduations/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))

	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	academicYearID, ok := parseOptionalUint64(w, r.URL.Query().Get("academic_year_id"), "academic_year_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, status, studentID, academicYearID)
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
			response.Error(w, http.StatusNotFound, "student graduation not found")
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

	item, err := buildStudentGraduation(req.StudentID, req.AcademicYearID, req.GraduationDate, req.Status, req.Notes)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "create", "student_graduation", created.ID, userID, req)

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

	item, err := buildStudentGraduation(req.StudentID, req.AcademicYearID, req.GraduationDate, req.Status, req.Notes)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "update", "student_graduation", id, userID, req)

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
			response.Error(w, http.StatusNotFound, "student graduation not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "delete", "student_graduation", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "student graduation deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "student graduation not found")
	case errors.Is(err, ErrStudentNotFound):
		response.Error(w, http.StatusBadRequest, "student not found")
	case errors.Is(err, ErrAcademicYearNotFound):
		response.Error(w, http.StatusBadRequest, "academic year not found")
	case errors.Is(err, ErrDuplicateScope):
		response.Error(w, http.StatusConflict, "student graduation already exists for the selected academic year")
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

func buildStudentGraduation(studentID, academicYearID uint64, graduationDate, status, notes string) (StudentGraduation, error) {
	if studentID == 0 {
		return StudentGraduation{}, fmt.Errorf("student_id is required")
	}
	if academicYearID == 0 {
		return StudentGraduation{}, fmt.Errorf("academic_year_id is required")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "graduated"
	}
	switch status {
	case "graduated", "pending", "cancelled":
	default:
		return StudentGraduation{}, fmt.Errorf("status must be one of graduated, pending, or cancelled")
	}

	var parsedGraduationDate *time.Time
	if value := strings.TrimSpace(graduationDate); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return StudentGraduation{}, fmt.Errorf("graduation_date must use YYYY-MM-DD format")
		}
		parsedGraduationDate = &parsed
	}

	return StudentGraduation{
		StudentID:      studentID,
		AcademicYearID: academicYearID,
		GraduationDate: parsedGraduationDate,
		Status:         status,
		Notes:          strings.TrimSpace(notes),
	}, nil
}
