package internshiplogs

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
	return &Handler{repo: NewRepository(db), auditLog: auditLog}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/industry-relations/internship-logs", h.List)
	mux.HandleFunc("POST /api/v1/industry-relations/internship-logs", h.Create)
	mux.HandleFunc("GET /api/v1/industry-relations/internship-logs/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/industry-relations/internship-logs/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/industry-relations/internship-logs/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	internshipID, ok := parseOptionalUint64(w, r.URL.Query().Get("internship_id"), "internship_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, internshipID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
		"meta":    map[string]any{"total": len(items)},
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
			response.Error(w, http.StatusNotFound, "internship log not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": item})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	item, err := buildInternshipLog(req.InternshipID, req.LogDate, req.Activity, req.Notes, req.SupervisorName)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		if errors.Is(err, ErrInternshipNotFound) {
			response.Error(w, http.StatusBadRequest, "internship not found")
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
	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "industry_relations", "create", "internship_log", created.ID, userID, req)
	response.JSON(w, http.StatusCreated, map[string]any{"success": true, "data": created})
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

	item, err := buildInternshipLog(req.InternshipID, req.LogDate, req.Activity, req.Notes, req.SupervisorName)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "internship log not found")
			return
		}
		if errors.Is(err, ErrInternshipNotFound) {
			response.Error(w, http.StatusBadRequest, "internship not found")
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
	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "industry_relations", "update", "internship_log", id, userID, req)
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": updated})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "internship log not found")
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
	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "industry_relations", "delete", "internship_log", id, userID, nil)
	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "internship log deleted successfully",
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

func buildInternshipLog(internshipID uint64, logDate, activity, notes, supervisorName string) (InternshipLog, error) {
	if internshipID == 0 {
		return InternshipLog{}, fmt.Errorf("internship_id is required")
	}

	activity = strings.TrimSpace(activity)
	if activity == "" {
		return InternshipLog{}, fmt.Errorf("activity is required")
	}

	logDate = strings.TrimSpace(logDate)
	if logDate == "" {
		return InternshipLog{}, fmt.Errorf("log_date is required")
	}

	parsedDate, err := time.Parse("2006-01-02", logDate)
	if err != nil {
		return InternshipLog{}, fmt.Errorf("log_date must use YYYY-MM-DD format")
	}

	return InternshipLog{
		InternshipID:   internshipID,
		LogDate:        parsedDate,
		Activity:       activity,
		Notes:          strings.TrimSpace(notes),
		SupervisorName: strings.TrimSpace(supervisorName),
	}, nil
}
