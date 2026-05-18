package studentenrollments

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"siakad/backend/internal/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		repo: NewRepository(db),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/student-affairs/enrollments", h.List)
	mux.HandleFunc("POST /api/v1/student-affairs/enrollments", h.Create)
	mux.HandleFunc("GET /api/v1/student-affairs/enrollments/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/student-affairs/enrollments/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/student-affairs/enrollments/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))

	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	classID, ok := parseOptionalUint64(w, r.URL.Query().Get("class_id"), "class_id")
	if !ok {
		return
	}
	academicYearID, ok := parseOptionalUint64(w, r.URL.Query().Get("academic_year_id"), "academic_year_id")
	if !ok {
		return
	}
	semesterID, ok := parseOptionalUint64(w, r.URL.Query().Get("semester_id"), "semester_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, status, studentID, classID, academicYearID, semesterID)
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
			response.Error(w, http.StatusNotFound, "student enrollment not found")
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

	item, err := buildStudentEnrollment(req.StudentID, req.ClassID, req.AcademicYearID, req.SemesterID, req.Status)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

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

	item, err := buildStudentEnrollment(req.StudentID, req.ClassID, req.AcademicYearID, req.SemesterID, req.Status)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

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
			response.Error(w, http.StatusNotFound, "student enrollment not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "student enrollment deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "student enrollment not found")
	case errors.Is(err, ErrStudentNotFound):
		response.Error(w, http.StatusBadRequest, "student not found")
	case errors.Is(err, ErrClassNotFound):
		response.Error(w, http.StatusBadRequest, "class not found")
	case errors.Is(err, ErrAcademicYearNotFound):
		response.Error(w, http.StatusBadRequest, "academic year not found")
	case errors.Is(err, ErrSemesterNotFound):
		response.Error(w, http.StatusBadRequest, "semester not found")
	case errors.Is(err, ErrSemesterAcademicYearMismatch):
		response.Error(w, http.StatusBadRequest, "semester does not belong to the selected academic year")
	case errors.Is(err, ErrClassAcademicYearMismatch):
		response.Error(w, http.StatusBadRequest, "class does not belong to the selected academic year")
	case errors.Is(err, ErrDuplicateScope):
		response.Error(w, http.StatusConflict, "student enrollment already exists for the selected academic year and semester")
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

func buildStudentEnrollment(studentID, classID, academicYearID, semesterID uint64, status string) (StudentEnrollment, error) {
	if studentID == 0 {
		return StudentEnrollment{}, fmt.Errorf("student_id is required")
	}
	if classID == 0 {
		return StudentEnrollment{}, fmt.Errorf("class_id is required")
	}
	if academicYearID == 0 {
		return StudentEnrollment{}, fmt.Errorf("academic_year_id is required")
	}
	if semesterID == 0 {
		return StudentEnrollment{}, fmt.Errorf("semester_id is required")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "active"
	}

	return StudentEnrollment{
		StudentID:      studentID,
		ClassID:        classID,
		AcademicYearID: academicYearID,
		SemesterID:     semesterID,
		Status:         status,
	}, nil
}
