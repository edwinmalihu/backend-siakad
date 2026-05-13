package homeroomassignments

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
	mux.HandleFunc("GET /api/v1/academic/homeroom-assignments", h.List)
	mux.HandleFunc("POST /api/v1/academic/homeroom-assignments", h.Create)
	mux.HandleFunc("GET /api/v1/academic/homeroom-assignments/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/academic/homeroom-assignments/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/academic/homeroom-assignments/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	teacherID, ok := parseOptionalUint64(w, r.URL.Query().Get("teacher_id"), "teacher_id")
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

	items, err := h.repo.List(r.Context(), search, teacherID, classID, academicYearID, semesterID)
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
			response.Error(w, http.StatusNotFound, "homeroom assignment not found")
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

	item, err := buildHomeroomAssignment(req.TeacherID, req.ClassID, req.AcademicYearID, req.SemesterID)
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

	item, err := buildHomeroomAssignment(req.TeacherID, req.ClassID, req.AcademicYearID, req.SemesterID)
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
			response.Error(w, http.StatusNotFound, "homeroom assignment not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "homeroom assignment deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "homeroom assignment not found")
	case errors.Is(err, ErrTeacherNotFound):
		response.Error(w, http.StatusBadRequest, "teacher not found")
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
	case errors.Is(err, ErrDuplicateClassScope):
		response.Error(w, http.StatusConflict, "homeroom assignment already exists for the selected class, academic year, and semester")
	case errors.Is(err, ErrTeacherAlreadyAssigned):
		response.Error(w, http.StatusConflict, "teacher is already assigned as homeroom teacher for the selected academic year and semester")
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

func buildHomeroomAssignment(teacherID, classID, academicYearID, semesterID uint64) (HomeroomAssignment, error) {
	if teacherID == 0 {
		return HomeroomAssignment{}, fmt.Errorf("teacher_id is required")
	}
	if classID == 0 {
		return HomeroomAssignment{}, fmt.Errorf("class_id is required")
	}
	if academicYearID == 0 {
		return HomeroomAssignment{}, fmt.Errorf("academic_year_id is required")
	}
	if semesterID == 0 {
		return HomeroomAssignment{}, fmt.Errorf("semester_id is required")
	}

	return HomeroomAssignment{
		TeacherID:      teacherID,
		ClassID:        classID,
		AcademicYearID: academicYearID,
		SemesterID:     semesterID,
	}, nil
}
