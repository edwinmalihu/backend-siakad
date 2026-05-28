package attendances

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
	mux.HandleFunc("GET /api/v1/student-affairs/attendances", h.List)
	mux.HandleFunc("POST /api/v1/student-affairs/attendances", h.Create)
	mux.HandleFunc("GET /api/v1/student-affairs/attendances/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/student-affairs/attendances/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/student-affairs/attendances/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))
	attendanceDate := strings.TrimSpace(r.URL.Query().Get("attendance_date"))
	if attendanceDate != "" {
		if _, err := time.Parse("2006-01-02", attendanceDate); err != nil {
			response.Error(w, http.StatusBadRequest, "attendance_date must use YYYY-MM-DD format")
			return
		}
	}

	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	classID, ok := parseOptionalUint64(w, r.URL.Query().Get("class_id"), "class_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, status, attendanceDate, studentID, classID)
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
			response.Error(w, http.StatusNotFound, "attendance not found")
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

	item, err := buildAttendance(req.StudentID, req.ClassID, req.AttendanceDate, req.Status, req.Notes, req.RecordedBy)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "create", "attendance", created.ID, userID, req)

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

	item, err := buildAttendance(req.StudentID, req.ClassID, req.AttendanceDate, req.Status, req.Notes, req.RecordedBy)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "update", "attendance", id, userID, req)

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
			response.Error(w, http.StatusNotFound, "attendance not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "delete", "attendance", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "attendance deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "attendance not found")
	case errors.Is(err, ErrStudentNotFound):
		response.Error(w, http.StatusBadRequest, "student not found")
	case errors.Is(err, ErrClassNotFound):
		response.Error(w, http.StatusBadRequest, "class not found")
	case errors.Is(err, ErrRecordedByNotFound):
		response.Error(w, http.StatusBadRequest, "recorded_by user not found")
	case errors.Is(err, ErrDuplicateScope):
		response.Error(w, http.StatusConflict, "attendance already exists for the selected student, class, and date")
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

func buildAttendance(studentID, classID uint64, attendanceDate, status, notes string, recordedBy uint64) (Attendance, error) {
	if studentID == 0 {
		return Attendance{}, fmt.Errorf("student_id is required")
	}
	if classID == 0 {
		return Attendance{}, fmt.Errorf("class_id is required")
	}
	if strings.TrimSpace(attendanceDate) == "" {
		return Attendance{}, fmt.Errorf("attendance_date is required")
	}
	parsedAttendanceDate, err := time.Parse("2006-01-02", strings.TrimSpace(attendanceDate))
	if err != nil {
		return Attendance{}, fmt.Errorf("attendance_date must use YYYY-MM-DD format")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "present", "sick", "excused", "absent", "late":
	default:
		return Attendance{}, fmt.Errorf("status must be one of present, sick, excused, absent, or late")
	}

	var recordedByValue *uint64
	if recordedBy > 0 {
		recordedByValue = &recordedBy
	}

	return Attendance{
		StudentID:      studentID,
		ClassID:        classID,
		AttendanceDate: parsedAttendanceDate,
		Status:         status,
		Notes:          strings.TrimSpace(notes),
		RecordedBy:     recordedByValue,
	}, nil
}
