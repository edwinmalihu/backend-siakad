package studentassessments

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
	mux.HandleFunc("GET /api/v1/academic/student-assessments", h.List)
	mux.HandleFunc("POST /api/v1/academic/student-assessments", h.Create)
	mux.HandleFunc("GET /api/v1/academic/student-assessments/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/academic/student-assessments/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/academic/student-assessments/{id}", h.Delete)
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
	assessmentComponentID, ok := parseOptionalUint64(w, r.URL.Query().Get("assessment_component_id"), "assessment_component_id")
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

	items, err := h.repo.List(r.Context(), search, studentID, classID, subjectID, assessmentComponentID, academicYearID, semesterID)
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
			response.Error(w, http.StatusNotFound, "student assessment not found")
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

	item, err := buildStudentAssessment(req.StudentID, req.ClassID, req.SubjectID, req.AssessmentComponentID, req.TeacherID, req.Score, req.AcademicYearID, req.SemesterID)
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
		case errors.Is(err, ErrAssessmentComponentNotFound):
			response.Error(w, http.StatusBadRequest, "assessment component not found")
		case errors.Is(err, ErrTeacherNotFound):
			response.Error(w, http.StatusBadRequest, "teacher not found")
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrSemesterNotFound):
			response.Error(w, http.StatusBadRequest, "semester not found")
		case errors.Is(err, ErrSemesterMismatch):
			response.Error(w, http.StatusBadRequest, "semester does not belong to the selected academic year")
		case errors.Is(err, ErrComponentMismatch):
			response.Error(w, http.StatusBadRequest, "assessment component does not belong to the selected subject and semester")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "student assessment already exists for this scope")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "create", "student_assessment", created.ID, userID, req)

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

	item, err := buildStudentAssessment(req.StudentID, req.ClassID, req.SubjectID, req.AssessmentComponentID, req.TeacherID, req.Score, req.AcademicYearID, req.SemesterID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "student assessment not found")
		case errors.Is(err, ErrStudentNotFound):
			response.Error(w, http.StatusBadRequest, "student not found")
		case errors.Is(err, ErrClassNotFound):
			response.Error(w, http.StatusBadRequest, "class not found")
		case errors.Is(err, ErrSubjectNotFound):
			response.Error(w, http.StatusBadRequest, "subject not found")
		case errors.Is(err, ErrAssessmentComponentNotFound):
			response.Error(w, http.StatusBadRequest, "assessment component not found")
		case errors.Is(err, ErrTeacherNotFound):
			response.Error(w, http.StatusBadRequest, "teacher not found")
		case errors.Is(err, ErrAcademicYearNotFound):
			response.Error(w, http.StatusBadRequest, "academic year not found")
		case errors.Is(err, ErrSemesterNotFound):
			response.Error(w, http.StatusBadRequest, "semester not found")
		case errors.Is(err, ErrSemesterMismatch):
			response.Error(w, http.StatusBadRequest, "semester does not belong to the selected academic year")
		case errors.Is(err, ErrComponentMismatch):
			response.Error(w, http.StatusBadRequest, "assessment component does not belong to the selected subject and semester")
		case errors.Is(err, ErrDuplicateScope):
			response.Error(w, http.StatusConflict, "student assessment already exists for this scope")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "update", "student_assessment", id, userID, req)

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
			response.Error(w, http.StatusNotFound, "student assessment not found")
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

	auditlogs.LogAuditWithID(r.Context(), r, h.auditLog, "academic", "delete", "student_assessment", id, userID, nil)

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "student assessment deleted successfully",
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

func buildStudentAssessment(studentID, classID, subjectID, assessmentComponentID, teacherID uint64, score float64, academicYearID, semesterID uint64) (StudentAssessment, error) {
	if studentID == 0 {
		return StudentAssessment{}, fmt.Errorf("student_id is required")
	}
	if classID == 0 {
		return StudentAssessment{}, fmt.Errorf("class_id is required")
	}
	if subjectID == 0 {
		return StudentAssessment{}, fmt.Errorf("subject_id is required")
	}
	if assessmentComponentID == 0 {
		return StudentAssessment{}, fmt.Errorf("assessment_component_id is required")
	}
	if teacherID == 0 {
		return StudentAssessment{}, fmt.Errorf("teacher_id is required")
	}
	if academicYearID == 0 {
		return StudentAssessment{}, fmt.Errorf("academic_year_id is required")
	}
	if semesterID == 0 {
		return StudentAssessment{}, fmt.Errorf("semester_id is required")
	}

	if score < 0 || score > 100 {
		return StudentAssessment{}, fmt.Errorf("score must be between 0 and 100")
	}

	return StudentAssessment{
		StudentID:            studentID,
		ClassID:              classID,
		SubjectID:            subjectID,
		AssessmentComponentID: assessmentComponentID,
		TeacherID:            teacherID,
		Score:                score,
		AcademicYearID:       academicYearID,
		SemesterID:           semesterID,
	}, nil
}
