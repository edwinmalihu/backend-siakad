package academicyears

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

const dateLayout = "2006-01-02"

type Handler struct {
	repo *Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		repo: NewRepository(db),
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/master/academic-years", h.List)
	mux.HandleFunc("POST /api/v1/master/academic-years", h.Create)
	mux.HandleFunc("GET /api/v1/master/academic-years/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/master/academic-years/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/master/academic-years/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))

	var isActive *bool
	if raw := strings.TrimSpace(r.URL.Query().Get("is_active")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "is_active must be a boolean")
			return
		}
		isActive = &parsed
	}

	items, err := h.repo.List(r.Context(), search, isActive)
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
			response.Error(w, http.StatusNotFound, "academic year not found")
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

	item, err := buildAcademicYear(req.Name, req.StartDate, req.EndDate, req.IsActive)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateName):
			response.Error(w, http.StatusConflict, "academic year name already exists")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
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

	item, err := buildAcademicYear(req.Name, req.StartDate, req.EndDate, req.IsActive)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			response.Error(w, http.StatusNotFound, "academic year not found")
		case errors.Is(err, ErrDuplicateName):
			response.Error(w, http.StatusConflict, "academic year name already exists")
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
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
			response.Error(w, http.StatusNotFound, "academic year not found")
			return
		}

		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "academic year deleted successfully",
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

func buildAcademicYear(name, startDate, endDate string, isActive bool) (AcademicYear, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return AcademicYear{}, fmt.Errorf("name is required")
	}

	start, err := time.Parse(dateLayout, strings.TrimSpace(startDate))
	if err != nil {
		return AcademicYear{}, fmt.Errorf("start_date must use YYYY-MM-DD format")
	}

	end, err := time.Parse(dateLayout, strings.TrimSpace(endDate))
	if err != nil {
		return AcademicYear{}, fmt.Errorf("end_date must use YYYY-MM-DD format")
	}

	if end.Before(start) {
		return AcademicYear{}, fmt.Errorf("end_date must be greater than or equal to start_date")
	}

	return AcademicYear{
		Name:      name,
		StartDate: start,
		EndDate:   end,
		IsActive:  isActive,
	}, nil
}
