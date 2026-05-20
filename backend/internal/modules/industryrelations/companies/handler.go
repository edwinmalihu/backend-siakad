package companies

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
	mux.HandleFunc("GET /api/v1/industry-relations/companies", h.List)
	mux.HandleFunc("POST /api/v1/industry-relations/companies", h.Create)
	mux.HandleFunc("GET /api/v1/industry-relations/companies/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/industry-relations/companies/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/industry-relations/companies/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	city := strings.TrimSpace(r.URL.Query().Get("city"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	categoryID, ok := parseOptionalUint64(w, r.URL.Query().Get("category_id"), "category_id")
	if !ok {
		return
	}
	items, err := h.repo.List(r.Context(), search, city, status, categoryID)
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
			response.Error(w, http.StatusNotFound, "company not found")
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
	item, err := buildCompany(req.CategoryID, req.Name, req.City, req.Address, req.ContactPerson, req.Phone, req.Email, req.Status)
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
	item, err := buildCompany(req.CategoryID, req.Name, req.City, req.Address, req.ContactPerson, req.Phone, req.Email, req.Status)
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
			response.Error(w, http.StatusNotFound, "company not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "company deleted successfully"})
}

func writeRepositoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		response.Error(w, http.StatusNotFound, "company not found")
	case errors.Is(err, ErrCategoryNotFound):
		response.Error(w, http.StatusBadRequest, "industry category not found")
	case errors.Is(err, ErrDuplicateName):
		response.Error(w, http.StatusConflict, "company name already exists")
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

func buildCompany(categoryID uint64, name, city, address, contactPerson, phone, email, status string) (Company, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Company{}, fmt.Errorf("name is required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
	}
	var categoryValue *uint64
	if categoryID > 0 {
		categoryValue = &categoryID
	}
	return Company{
		CategoryID:    categoryValue,
		Name:          name,
		City:          strings.TrimSpace(city),
		Address:       strings.TrimSpace(address),
		ContactPerson: strings.TrimSpace(contactPerson),
		Phone:         strings.TrimSpace(phone),
		Email:         strings.TrimSpace(email),
		Status:        status,
	}, nil
}
