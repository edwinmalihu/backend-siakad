package semesters

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	mux.HandleFunc("GET /api/v1/master/semesters", h.List)
	mux.HandleFunc("POST /api/v1/master/semesters", h.Create)
	mux.HandleFunc("GET /api/v1/master/semesters/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/master/semesters/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/master/semesters/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	var academicYearID uint64
	if raw := strings.TrimSpace(r.URL.Query().Get("academic_year_id")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "academic_year_id must be a valid integer")
			return
		}
		academicYearID = parsed
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

	items, err := h.repo.List(r.Context(), search, academicYearID, isActive)
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
			response.Error(w, http.StatusNotFound, "semester not found")
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

	item, err := buildSemester(req.AcademicYearID, req.Name, req.Code, req.IsActive)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		switch {
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "semester code already exists in the selected academic year")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var userID *uint64
	if user != nil {
		uid := user.UserID
		userID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "master", "create", "semester", created.ID, userID, req)

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

	item, err := buildSemester(req.AcademicYearID, req.Name, req.Code, req.IsActive)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "semester not found")
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "semester code already exists in the selected academic year")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	user := auth.GetUserFromContext(r.Context())
	var userID *uint64
	if user != nil {
		uid := user.UserID
		userID = &uid
	}

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "master", "update", "semester", updated.ID, userID, req)

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
			response.Error(w, http.StatusNotFound, "semester not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "master", "delete", "semester", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "semester deleted successfully",
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

func buildSemester(academicYearID uint64, name, code string, isActive bool) (Semester, error) {
	if academicYearID == 0 {
		return Semester{}, fmt.Errorf("academic_year_id is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return Semester{}, fmt.Errorf("name is required")
	}

	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return Semester{}, fmt.Errorf("code is required")
	}

	return Semester{
		AcademicYearID: academicYearID,
		Name:           name,
		Code:           code,
		IsActive:       isActive,
	}, nil
}
