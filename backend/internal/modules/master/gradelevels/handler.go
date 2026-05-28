package gradelevels

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
	mux.HandleFunc("GET /api/v1/master/grade-levels", h.List)
	mux.HandleFunc("POST /api/v1/master/grade-levels", h.Create)
	mux.HandleFunc("GET /api/v1/master/grade-levels/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/master/grade-levels/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/master/grade-levels/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	items, err := h.repo.List(r.Context(), search)
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
			response.Error(w, http.StatusNotFound, "grade level not found")
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

	item, err := buildGradeLevel(req.Code, req.Name, req.SortOrder)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateCode):
			response.Error(w, http.StatusConflict, "grade level code already exists")
		case errors.Is(err, ErrDuplicateName):
			response.Error(w, http.StatusConflict, "grade level name already exists")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "master", "create", "grade_level", created.ID, userID, req)

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

	item, err := buildGradeLevel(req.Code, req.Name, req.SortOrder)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "grade level not found")
		case errors.Is(err, ErrDuplicateCode):
			response.Error(w, http.StatusConflict, "grade level code already exists")
		case errors.Is(err, ErrDuplicateName):
			response.Error(w, http.StatusConflict, "grade level name already exists")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "master", "update", "grade_level", updated.ID, userID, req)

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
			response.Error(w, http.StatusNotFound, "grade level not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "master", "delete", "grade_level", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "grade level deleted successfully",
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

func buildGradeLevel(code, name string, sortOrder int) (GradeLevel, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return GradeLevel{}, fmt.Errorf("code is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return GradeLevel{}, fmt.Errorf("name is required")
	}

	if sortOrder < 0 {
		return GradeLevel{}, fmt.Errorf("sort_order must be greater than or equal to 0")
	}

	return GradeLevel{
		Code:      code,
		Name:      name,
		SortOrder: sortOrder,
	}, nil
}
