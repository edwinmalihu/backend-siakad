package alumni

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
	return &Handler{repo: NewRepository(db)}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/industry-relations/alumni", h.List)
	mux.HandleFunc("POST /api/v1/industry-relations/alumni", h.Create)
	mux.HandleFunc("GET /api/v1/industry-relations/alumni/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/industry-relations/alumni/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/industry-relations/alumni/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	currentActivity := strings.TrimSpace(r.URL.Query().Get("current_activity"))
	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	graduationYear, ok := parseOptionalInt(w, r.URL.Query().Get("graduation_year"), "graduation_year")
	if !ok {
		return
	}
	items, err := h.repo.List(r.Context(), search, currentActivity, studentID, graduationYear)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": items, "meta": map[string]any{"total": len(items)}})
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "alumnus not found")
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
	item, err := buildAlumnus(req.StudentID, req.GraduationYear, req.CurrentActivity, req.CompanyName, req.CollegeName, req.Phone, req.Email)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}
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
	item, err := buildAlumnus(req.StudentID, req.GraduationYear, req.CurrentActivity, req.CompanyName, req.CollegeName, req.Phone, req.Email)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": updated})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "alumnus not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "alumnus deleted successfully"})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "alumnus not found")
	case errors.Is(err, ErrStudentNotFound):
		response.Error(w, http.StatusBadRequest, "student not found")
	case errors.Is(err, ErrDuplicateStudent):
		response.Error(w, http.StatusConflict, "student is already registered as alumni")
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

func buildAlumnus(studentID uint64, graduationYear int, currentActivity, companyName, collegeName, phone, email string) (Alumnus, error) {
	if studentID == 0 {
		return Alumnus{}, fmt.Errorf("student_id is required")
	}
	if graduationYear < 2000 || graduationYear > 2200 {
		return Alumnus{}, fmt.Errorf("graduation_year must be between 2000 and 2200")
	}
	return Alumnus{
		StudentID:       studentID,
		GraduationYear:  graduationYear,
		CurrentActivity: strings.TrimSpace(currentActivity),
		CompanyName:     strings.TrimSpace(companyName),
		CollegeName:     strings.TrimSpace(collegeName),
		Phone:           strings.TrimSpace(phone),
		Email:           strings.TrimSpace(email),
	}, nil
}
