package classes

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
	mux.HandleFunc("GET /api/v1/master/classes", h.List)
	mux.HandleFunc("POST /api/v1/master/classes", h.Create)
	mux.HandleFunc("GET /api/v1/master/classes/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/master/classes/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/master/classes/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	academicYearID, ok := parseOptionalUint64(w, r.URL.Query().Get("academic_year_id"), "academic_year_id")
	if !ok {
		return
	}
	departmentID, ok := parseOptionalUint64(w, r.URL.Query().Get("department_id"), "department_id")
	if !ok {
		return
	}
	gradeLevelID, ok := parseOptionalUint64(w, r.URL.Query().Get("grade_level_id"), "grade_level_id")
	if !ok {
		return
	}

	var isActive *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("is_active")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "is_active must be a boolean")
			return
		}
		isActive = &parsed
	}

	items, err := h.repo.List(r.Context(), search, academicYearID, departmentID, gradeLevelID, isActive)
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
			response.Error(w, http.StatusNotFound, "class not found")
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

	item, err := buildClass(req.AcademicYearID, req.DepartmentID, req.GradeLevelID, req.Name, req.IsActive)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		switch {
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrDepartmentNotFound):
			response.Error(w, http.StatusBadRequest, "department not found")
		case errors.Is(err, ErrGradeLevelNotFound):
			response.Error(w, http.StatusBadRequest, "grade level not found")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "class already exists in the selected academic year, department, and grade level")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
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

	item, err := buildClass(req.AcademicYearID, req.DepartmentID, req.GradeLevelID, req.Name, req.IsActive)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "class not found")
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrDepartmentNotFound):
			response.Error(w, http.StatusBadRequest, "department not found")
		case errors.Is(err, ErrGradeLevelNotFound):
			response.Error(w, http.StatusBadRequest, "grade level not found")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "class already exists in the selected academic year, department, and grade level")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
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
			response.Error(w, http.StatusNotFound, "class not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "class deleted successfully",
	})
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

func buildClass(academicYearID, departmentID, gradeLevelID uint64, name string, isActive bool) (Class, error) {
	if academicYearID == 0 {
		return Class{}, fmt.Errorf("academic_year_id is required")
	}
	if departmentID == 0 {
		return Class{}, fmt.Errorf("department_id is required")
	}
	if gradeLevelID == 0 {
		return Class{}, fmt.Errorf("grade_level_id is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Class{}, fmt.Errorf("name is required")
	}

	return Class{
		AcademicYearID: academicYearID,
		DepartmentID:   departmentID,
		GradeLevelID:   gradeLevelID,
		Name:           name,
		IsActive:       isActive,
	}, nil
}
