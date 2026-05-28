package students

import (
	"context"
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
	mux.HandleFunc("GET /api/v1/student-affairs/students", h.List)
	mux.HandleFunc("POST /api/v1/student-affairs/students", h.Create)
	mux.HandleFunc("GET /api/v1/student-affairs/students/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/student-affairs/students/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/student-affairs/students/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	gender := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("gender")))
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))

	if gender != "" && gender != "male" && gender != "female" {
		response.Error(w, http.StatusBadRequest, "gender must be either male or female")
		return
	}

	entryYear, ok := parseOptionalInt(w, r.URL.Query().Get("entry_year"), "entry_year")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, gender, status, entryYear)
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
			response.Error(w, http.StatusNotFound, "student not found")
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

	item, err := buildStudent(req.NIS, req.NISN, req.FullName, req.Gender, req.BirthPlace, req.BirthDate, req.Address, req.Phone, req.EntryYear, req.Status)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	h.logAudit(r.Context(), r, "create", "student", created.ID, req)

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

	item, err := buildStudent(req.NIS, req.NISN, req.FullName, req.Gender, req.BirthPlace, req.BirthDate, req.Address, req.Phone, req.EntryYear, req.Status)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

	h.logAudit(r.Context(), r, "update", "student", id, req)

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
			response.Error(w, http.StatusNotFound, "student not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logAudit(r.Context(), r, "delete", "student", id, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "student deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "student not found")
	case errors.Is(err, ErrDuplicateNIS):
		response.Error(w, http.StatusConflict, "student nis already exists")
	case errors.Is(err, ErrDuplicateNISN):
		response.Error(w, http.StatusConflict, "student nisn already exists")
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

func parseOptionalInt(w http.ResponseWriter, raw, field string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, true
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s must be a valid integer", field))
		return 0, false
	}
	return parsed, true
}

func buildStudent(nis, nisn, fullName, gender, birthPlace, birthDate, address, phone string, entryYear int, status string) (Student, error) {
	nis = strings.ToUpper(strings.TrimSpace(nis))
	if nis == "" {
		return Student{}, fmt.Errorf("nis is required")
	}

	nisn = strings.TrimSpace(nisn)
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return Student{}, fmt.Errorf("full_name is required")
	}

	gender = strings.ToLower(strings.TrimSpace(gender))
	if gender != "male" && gender != "female" {
		return Student{}, fmt.Errorf("gender must be either male or female")
	}

	if entryYear < 1901 || entryYear > 2155 {
		return Student{}, fmt.Errorf("entry_year must be between 1901 and 2155")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "active"
	}

	var parsedBirthDate *time.Time
	if value := strings.TrimSpace(birthDate); value != "" {
		parsed, err := time.Parse("2006-01-02", value)
		if err != nil {
			return Student{}, fmt.Errorf("birth_date must use YYYY-MM-DD format")
		}
		parsedBirthDate = &parsed
	}

	return Student{
		NIS:        nis,
		NISN:       nisn,
		FullName:   fullName,
		Gender:     gender,
		BirthPlace: strings.TrimSpace(birthPlace),
		BirthDate:  parsedBirthDate,
		Address:    strings.TrimSpace(address),
		Phone:      strings.TrimSpace(phone),
		EntryYear:  entryYear,
		Status:     status,
	}, nil
}

func (h *Handler) logAudit(ctx context.Context, r *http.Request, action, entityType string, entityID uint64, payload interface{}) {
	if h.auditLog == nil {
		return
	}

	user := auth.GetUserFromContext(ctx)
	var userID *uint64
	if user != nil {
		uid := user.UserID
		userID = &uid
	}

	payloadJSON := ""
	if payload != nil {
		if bytes, err := json.Marshal(payload); err == nil {
			payloadJSON = string(bytes)
		}
	}

	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ipAddress = strings.TrimSpace(parts[0])
		}
	}

	auditLog := &auditlogs.AuditLog{
		UserID:      userID,
		Module:      "student_affairs",
		Action:      action,
		EntityType:  entityType,
		EntityID:    &entityID,
		PayloadJSON: payloadJSON,
		IPAddress:   ipAddress,
	}

	_ = h.auditLog.Create(ctx, auditLog)
}
