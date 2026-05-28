package extracurriculars

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
	mux.HandleFunc("GET /api/v1/student-affairs/extracurriculars", h.List)
	mux.HandleFunc("POST /api/v1/student-affairs/extracurriculars", h.Create)
	mux.HandleFunc("GET /api/v1/student-affairs/extracurriculars/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/student-affairs/extracurriculars/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/student-affairs/extracurriculars/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	coachTeacherID, ok := parseOptionalUint64(w, r.URL.Query().Get("coach_teacher_id"), "coach_teacher_id")
	if !ok {
		return
	}
	isActive, ok := parseOptionalBool(w, r.URL.Query().Get("is_active"), "is_active")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, coachTeacherID, isActive)
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
			response.Error(w, http.StatusNotFound, "extracurricular not found")
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
	item, err := buildExtracurricular(req.CoachTeacherID, req.Name, req.Description, req.IsActive)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "create", "extracurricular", created.ID, userID, req)

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
	item, err := buildExtracurricular(req.CoachTeacherID, req.Name, req.Description, req.IsActive)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "update", "extracurricular", id, userID, req)

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
			response.Error(w, http.StatusNotFound, "extracurricular not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "student_affairs", "delete", "extracurricular", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "extracurricular deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "extracurricular not found")
	case errors.Is(err, ErrCoachTeacherNotFound):
		response.Error(w, http.StatusBadRequest, "coach teacher not found")
	case errors.Is(err, ErrDuplicateName):
		response.Error(w, http.StatusConflict, "extracurricular name already exists")
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

func parseOptionalBool(w http.ResponseWriter, raw, field string) (*bool, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, true
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s must be a valid boolean", field))
		return nil, false
	}
	return &parsed, true
}

func buildExtracurricular(coachTeacherID uint64, name, description string, isActive bool) (Extracurricular, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Extracurricular{}, fmt.Errorf("name is required")
	}
	var coachTeacherValue *uint64
	if coachTeacherID > 0 {
		coachTeacherValue = &coachTeacherID
	}
	return Extracurricular{
		CoachTeacherID: coachTeacherValue,
		Name:           name,
		Description:    strings.TrimSpace(description),
		IsActive:       isActive,
	}, nil
}
