package studentgrades

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
	mux.HandleFunc("GET /api/v1/academic/student-grades", h.List)
	mux.HandleFunc("POST /api/v1/academic/student-grades", h.Create)
	mux.HandleFunc("GET /api/v1/academic/student-grades/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/academic/student-grades/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/academic/student-grades/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	studentID, ok := parseOptionalUint64(w, r.URL.Query().Get("student_id"), "student_id")
	if !ok {
		return
	}
	classID, ok := parseOptionalUint64(w, r.URL.Query().Get("class_id"), "class_id")
	if !ok {
		return
	}
	subjectID, ok := parseOptionalUint64(w, r.URL.Query().Get("subject_id"), "subject_id")
	if !ok {
		return
	}
	academicYearID, ok := parseOptionalUint64(w, r.URL.Query().Get("academic_year_id"), "academic_year_id")
	if !ok {
		return
	}
	semesterID, ok := parseOptionalUint64(w, r.URL.Query().Get("semester_id"), "semester_id")
	if !ok {
		return
	}

	items, err := h.repo.List(r.Context(), search, studentID, classID, subjectID, academicYearID, semesterID)
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
			response.Error(w, http.StatusNotFound, "student grade not found")
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

	item, err := buildStudentGrade(req.StudentID, req.ClassID, req.SubjectID, req.AcademicYearID, req.SemesterID, req.FinalScore, req.GradeLetter, req.Predicate)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		switch {
		case errors.Is(err, ErrStudentNotFound):
			response.Error(w, http.StatusBadRequest, "student not found")
		case errors.Is(err, ErrClassNotFound):
			response.Error(w, http.StatusBadRequest, "class not found")
		case errors.Is(err, ErrSubjectNotFound):
			response.Error(w, http.StatusBadRequest, "subject not found")
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrSemesterNotFound):
			response.Error(w, http.StatusBadRequest, "semester not found")
		case errors.Is(err, ErrSemesterMismatch):
			response.Error(w, http.StatusBadRequest, "semester does not belong to the selected academic year")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "student grade already exists for this scope")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "create", "student_grade", created.ID, userID, req)

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

	item, err := buildStudentGrade(req.StudentID, req.ClassID, req.SubjectID, req.AcademicYearID, req.SemesterID, req.FinalScore, req.GradeLetter, req.Predicate)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "student grade not found")
		case errors.Is(err, ErrStudentNotFound):
			response.Error(w, http.StatusBadRequest, "student not found")
		case errors.Is(err, ErrClassNotFound):
			response.Error(w, http.StatusBadRequest, "class not found")
		case errors.Is(err, ErrSubjectNotFound):
			response.Error(w, http.StatusBadRequest, "subject not found")
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrSemesterNotFound):
			response.Error(w, http.StatusBadRequest, "semester not found")
		case errors.Is(err, ErrSemesterMismatch):
			response.Error(w, http.StatusBadRequest, "semester does not belong to the selected academic year")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "student grade already exists for this scope")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "update", "student_grade", id, userID, req)

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
			response.Error(w, http.StatusNotFound, "student grade not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "delete", "student_grade", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "student grade deleted successfully",
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

func buildStudentGrade(studentID, classID, subjectID, academicYearID, semesterID uint64, finalScore float64, gradeLetter, predicate string) (StudentGrade, error) {
	if studentID == 0 {
		return StudentGrade{}, fmt.Errorf("student_id is required")
	}
	if classID == 0 {
		return StudentGrade{}, fmt.Errorf("class_id is required")
	}
	if subjectID == 0 {
		return StudentGrade{}, fmt.Errorf("subject_id is required")
	}
	if academicYearID == 0 {
		return StudentGrade{}, fmt.Errorf("academic_year_id is required")
	}
	if semesterID == 0 {
		return StudentGrade{}, fmt.Errorf("semester_id is required")
	}

	if finalScore < 0 || finalScore > 100 {
		return StudentGrade{}, fmt.Errorf("final_score must be between 0 and 100")
	}

	return StudentGrade{
		StudentID:      studentID,
		ClassID:        classID,
		SubjectID:      subjectID,
		AcademicYearID: academicYearID,
		SemesterID:     semesterID,
		FinalScore:     finalScore,
		GradeLetter:    strings.TrimSpace(gradeLetter),
		Predicate:      strings.TrimSpace(predicate),
	}, nil
}
