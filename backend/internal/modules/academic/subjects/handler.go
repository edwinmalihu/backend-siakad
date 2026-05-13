package subjects

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
	mux.HandleFunc("GET /api/v1/academic/subjects", h.List)
	mux.HandleFunc("POST /api/v1/academic/subjects", h.Create)
	mux.HandleFunc("GET /api/v1/academic/subjects/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/academic/subjects/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/academic/subjects/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	departmentID, ok := parseOptionalUint64(w, r.URL.Query().Get("department_id"), "department_id")
	if !ok {
		return
	}
	gradeLevelID, ok := parseOptionalUint64(w, r.URL.Query().Get("grade_level_id"), "grade_level_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, departmentID, gradeLevelID)
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
			response.Error(w, http.StatusNotFound, "subject not found")
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

	item, err := buildSubject(req.DepartmentID, req.GradeLevelID, req.Code, req.Name, req.SubjectType, req.KKM)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		switch {
		case errors.Is(err, ErrDepartmentNotFound):
			response.Error(w, http.StatusBadRequest, "department not found")
		case errors.Is(err, ErrGradeLevelNotFound):
			response.Error(w, http.StatusBadRequest, "grade level not found")
		case errors.Is(err, ErrDuplicateCode):
			response.Error(w, http.StatusConflict, "subject code already exists")
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

	item, err := buildSubject(req.DepartmentID, req.GradeLevelID, req.Code, req.Name, req.SubjectType, req.KKM)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "subject not found")
		case errors.Is(err, ErrDepartmentNotFound):
			response.Error(w, http.StatusBadRequest, "department not found")
		case errors.Is(err, ErrGradeLevelNotFound):
			response.Error(w, http.StatusBadRequest, "grade level not found")
		case errors.Is(err, ErrDuplicateCode):
			response.Error(w, http.StatusConflict, "subject code already exists")
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
			response.Error(w, http.StatusNotFound, "subject not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "subject deleted successfully",
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

func buildSubject(departmentID, gradeLevelID uint64, code, name, subjectType string, kkm *float64) (Subject, error) {
	if departmentID == 0 {
		return Subject{}, fmt.Errorf("department_id is required")
	}
	if gradeLevelID == 0 {
		return Subject{}, fmt.Errorf("grade_level_id is required")
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return Subject{}, fmt.Errorf("code is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Subject{}, fmt.Errorf("name is required")
	}

	subjectType = strings.TrimSpace(subjectType)
	if kkm != nil && (*kkm < 0 || *kkm > 100) {
		return Subject{}, fmt.Errorf("kkm must be between 0 and 100")
	}

	return Subject{
		DepartmentID: departmentID,
		GradeLevelID: gradeLevelID,
		Code:         code,
		Name:         name,
		SubjectType:  subjectType,
		KKM:          kkm,
	}, nil
}
