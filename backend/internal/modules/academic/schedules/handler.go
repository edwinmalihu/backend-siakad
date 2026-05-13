package schedules

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
	return &Handler{
		repo: NewRepository(db),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/academic/schedules", h.List)
	mux.HandleFunc("POST /api/v1/academic/schedules", h.Create)
	mux.HandleFunc("GET /api/v1/academic/schedules/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/academic/schedules/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/academic/schedules/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	classID, ok := parseOptionalUint64(w, r.URL.Query().Get("class_id"), "class_id")
	if !ok {
		return
	}
	teacherID, ok := parseOptionalUint64(w, r.URL.Query().Get("teacher_id"), "teacher_id")
	if !ok {
		return
	}
	roomID, ok := parseOptionalUint64(w, r.URL.Query().Get("room_id"), "room_id")
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

	var dayOfWeek *uint8
	if raw := strings.TrimSpace(r.URL.Query().Get("day_of_week")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 8)
		if err != nil || parsed < 1 || parsed > 7 {
			response.Error(w, http.StatusBadRequest, "day_of_week must be between 1 and 7")
			return
		}

		value := uint8(parsed)
		dayOfWeek = &value
	}

	items, err := h.repo.List(r.Context(), search, classID, teacherID, roomID, academicYearID, semesterID, dayOfWeek)
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
			response.Error(w, http.StatusNotFound, "schedule not found")
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

	item, err := buildSchedule(req.ClassID, req.SubjectID, req.TeacherID, req.RoomID, req.AcademicYearID, req.SemesterID, req.DayOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

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

	item, err := buildSchedule(req.ClassID, req.SubjectID, req.TeacherID, req.RoomID, req.AcademicYearID, req.SemesterID, req.DayOfWeek, req.StartTime, req.EndTime)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		writeRepositoryError(w, err)
		return
	}

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
			response.Error(w, http.StatusNotFound, "schedule not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "schedule deleted successfully",
	})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "schedule not found")
	case errors.Is(err, ErrAcademicYearNotFound):
		response.Error(w, http.StatusBadRequest, "academic year not found")
	case errors.Is(err, ErrSemesterNotFound):
		response.Error(w, http.StatusBadRequest, "semester not found")
	case errors.Is(err, ErrClassNotFound):
		response.Error(w, http.StatusBadRequest, "class not found")
	case errors.Is(err, ErrSubjectNotFound):
		response.Error(w, http.StatusBadRequest, "subject not found")
	case errors.Is(err, ErrTeacherNotFound):
		response.Error(w, http.StatusBadRequest, "teacher not found")
	case errors.Is(err, ErrRoomNotFound):
		response.Error(w, http.StatusBadRequest, "room not found")
	case errors.Is(err, ErrSemesterAcademicYearMismatch):
		response.Error(w, http.StatusBadRequest, "semester does not belong to the selected academic year")
	case errors.Is(err, ErrClassAcademicYearMismatch):
		response.Error(w, http.StatusBadRequest, "class does not belong to the selected academic year")
	case errors.Is(err, ErrSubjectScopeMismatch):
		response.Error(w, http.StatusBadRequest, "subject does not match the selected class scope")
	case errors.Is(err, ErrClassScheduleConflict):
		response.Error(w, http.StatusConflict, "class schedule conflicts with another schedule")
	case errors.Is(err, ErrTeacherScheduleConflict):
		response.Error(w, http.StatusConflict, "teacher schedule conflicts with another schedule")
	case errors.Is(err, ErrRoomScheduleConflict):
		response.Error(w, http.StatusConflict, "room schedule conflicts with another schedule")
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

func buildSchedule(classID, subjectID, teacherID uint64, roomID *uint64, academicYearID, semesterID uint64, dayOfWeek uint8, startTime, endTime string) (Schedule, error) {
	if classID == 0 {
		return Schedule{}, fmt.Errorf("class_id is required")
	}
	if subjectID == 0 {
		return Schedule{}, fmt.Errorf("subject_id is required")
	}
	if teacherID == 0 {
		return Schedule{}, fmt.Errorf("teacher_id is required")
	}
	if academicYearID == 0 {
		return Schedule{}, fmt.Errorf("academic_year_id is required")
	}
	if semesterID == 0 {
		return Schedule{}, fmt.Errorf("semester_id is required")
	}
	if dayOfWeek < 1 || dayOfWeek > 7 {
		return Schedule{}, fmt.Errorf("day_of_week must be between 1 and 7")
	}

	normalizedStart, startValue, err := normalizeClock(startTime)
	if err != nil {
		return Schedule{}, fmt.Errorf("start_time %w", err)
	}
	normalizedEnd, endValue, err := normalizeClock(endTime)
	if err != nil {
		return Schedule{}, fmt.Errorf("end_time %w", err)
	}
	if !startValue.Before(endValue) {
		return Schedule{}, fmt.Errorf("start_time must be earlier than end_time")
	}

	if roomID != nil && *roomID == 0 {
		roomID = nil
	}

	return Schedule{
		ClassID:        classID,
		SubjectID:      subjectID,
		TeacherID:      teacherID,
		RoomID:         roomID,
		AcademicYearID: academicYearID,
		SemesterID:     semesterID,
		DayOfWeek:      dayOfWeek,
		StartTime:      normalizedStart,
		EndTime:        normalizedEnd,
	}, nil
}

func normalizeClock(value string) (string, time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", time.Time{}, fmt.Errorf("is required")
	}

	layouts := []string{"15:04:05", "15:04"}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.Format("15:04:05"), parsed, nil
		}
	}

	return "", time.Time{}, fmt.Errorf("must be a valid time in HH:MM or HH:MM:SS format")
}
