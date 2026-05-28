package teachers

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
	mux.HandleFunc("GET /api/v1/academic/teachers", h.List)
	mux.HandleFunc("POST /api/v1/academic/teachers", h.Create)
	mux.HandleFunc("GET /api/v1/academic/teachers/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/academic/teachers/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/academic/teachers/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	gender := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("gender")))
	status := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))

	if gender != "" && gender != "male" && gender != "female" {
		response.Error(w, http.StatusBadRequest, "gender must be either male or female")
		return
	}

	items, err := h.repo.List(r.Context(), search, gender, status)
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
			response.Error(w, http.StatusNotFound, "teacher not found")
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

	item, err := buildTeacher(req.NIP, req.NUPTK, req.FullName, req.Gender, req.Address, req.Phone, req.Email, req.EmploymentStatus, req.Position, req.PhotoURL, req.Status)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "create", "teacher", created.ID, userID, req)

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

	item, err := buildTeacher(req.NIP, req.NUPTK, req.FullName, req.Gender, req.Address, req.Phone, req.Email, req.EmploymentStatus, req.Position, req.PhotoURL, req.Status)
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "update", "teacher", id, userID, req)

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
			response.Error(w, http.StatusNotFound, "teacher not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "delete", "teacher", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "teacher deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "teacher not found")
	case errors.Is(err, ErrDuplicateNIP):
		response.Error(w, http.StatusConflict, "teacher nip already exists")
	case errors.Is(err, ErrDuplicateNUPTK):
		response.Error(w, http.StatusConflict, "teacher nuptk already exists")
	case errors.Is(err, ErrDuplicateEmail):
		response.Error(w, http.StatusConflict, "teacher email already exists")
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

func buildTeacher(nip, nuptk, fullName, gender, address, phone, email, employmentStatus, position, photoURL, status string) (Teacher, error) {
	nip = strings.ToUpper(strings.TrimSpace(nip))
	nuptk = strings.ToUpper(strings.TrimSpace(nuptk))
	fullName = strings.TrimSpace(fullName)
	gender = strings.ToLower(strings.TrimSpace(gender))
	address = strings.TrimSpace(address)
	phone = strings.TrimSpace(phone)
	email = strings.ToLower(strings.TrimSpace(email))
	employmentStatus = strings.TrimSpace(employmentStatus)
	position = strings.TrimSpace(position)
	photoURL = strings.TrimSpace(photoURL)
	status = strings.ToLower(strings.TrimSpace(status))

	if fullName == "" {
		return Teacher{}, fmt.Errorf("full_name is required")
	}
	if gender != "" && gender != "male" && gender != "female" {
		return Teacher{}, fmt.Errorf("gender must be either male or female")
	}
	if email != "" && !strings.Contains(email, "@") {
		return Teacher{}, fmt.Errorf("email must be a valid email address")
	}
	if status == "" {
		status = "active"
	}

	return Teacher{
		NIP:              nip,
		NUPTK:            nuptk,
		FullName:         fullName,
		Gender:           gender,
		Address:          address,
		Phone:            phone,
		Email:            email,
		EmploymentStatus: employmentStatus,
		Position:         position,
		PhotoURL:         photoURL,
		Status:           status,
	}, nil
}
