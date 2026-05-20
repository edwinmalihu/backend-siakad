package internships

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"siakad/backend/internal/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: NewRepository(db)}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/industry-relations/internships", h.List)
	mux.HandleFunc("POST /api/v1/industry-relations/internships", h.Create)
	mux.HandleFunc("GET /api/v1/industry-relations/internships/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/industry-relations/internships/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/industry-relations/internships/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	companyID, ok := parseOptionalUint64(w, r.URL.Query().Get("company_id"), "company_id")
	if !ok {
		return
	}
	academicYearID, ok := parseOptionalUint64(w, r.URL.Query().Get("academic_year_id"), "academic_year_id")
	if !ok {
		return
	}
	items, err := h.repo.List(r.Context(), search, status, studentID, companyID, academicYearID)
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
			response.Error(w, http.StatusNotFound, "internship not found")
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
	item, err := buildInternship(req.StudentID, req.CompanyID, req.AcademicYearID, req.StartDate, req.EndDate, req.MentorName, req.Status)
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
	item, err := buildInternship(req.StudentID, req.CompanyID, req.AcademicYearID, req.StartDate, req.EndDate, req.MentorName, req.Status)
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
			response.Error(w, http.StatusNotFound, "internship not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "internship deleted successfully"})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "internship not found")
	case errors.Is(err, ErrStudentNotFound):
		response.Error(w, http.StatusBadRequest, "student not found")
	case errors.Is(err, ErrCompanyNotFound):
		response.Error(w, http.StatusBadRequest, "company not found")
	case errors.Is(err, ErrAcademicYearNotFound):
		response.Error(w, http.StatusBadRequest, "academic year not found")
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

func buildInternship(studentID, companyID, academicYearID uint64, startDate, endDate, mentorName, status string) (Internship, error) {
	if studentID == 0 {
		return Internship{}, fmt.Errorf("student_id is required")
	}
	if companyID == 0 {
		return Internship{}, fmt.Errorf("company_id is required")
	}
	if academicYearID == 0 {
		return Internship{}, fmt.Errorf("academic_year_id is required")
	}
	parsedStartDate, err := parseOptionalDate(startDate)
	if err != nil {
		return Internship{}, fmt.Errorf("start_date must use YYYY-MM-DD format")
	}
	parsedEndDate, err := parseOptionalDate(endDate)
	if err != nil {
		return Internship{}, fmt.Errorf("end_date must use YYYY-MM-DD format")
	}
	if parsedStartDate != nil && parsedEndDate != nil && parsedEndDate.Before(*parsedStartDate) {
		return Internship{}, fmt.Errorf("end_date must be on or after start_date")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "planned"
	}
	return Internship{
		StudentID:      studentID,
		CompanyID:      companyID,
		AcademicYearID: academicYearID,
		StartDate:      parsedStartDate,
		EndDate:        parsedEndDate,
		MentorName:     strings.TrimSpace(mentorName),
		Status:         status,
	}, nil
}

func parseOptionalDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
