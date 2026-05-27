package permissions

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	mux.HandleFunc("GET /api/v1/permissions", h.List)
	mux.HandleFunc("POST /api/v1/permissions", h.Create)
	mux.HandleFunc("GET /api/v1/permissions/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/permissions/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/permissions/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	items, err := h.repo.List(r.Context(), search)
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
			response.Error(w, http.StatusNotFound, "permission not found")
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
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(req.Code)
	if name == "" {
		response.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}
	item, err := h.repo.Create(r.Context(), Permission{Name: name, Code: code, Description: strings.TrimSpace(req.Description)})
	if err != nil {
		if errors.Is(err, ErrDuplicateName) {
			response.Error(w, http.StatusConflict, "permission name already exists")
			return
		}
		if errors.Is(err, ErrDuplicateCode) {
			response.Error(w, http.StatusConflict, "permission code already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, map[string]any{"success": true, "data": item})
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
	name := strings.TrimSpace(req.Name)
	code := strings.TrimSpace(req.Code)
	if name == "" {
		response.Error(w, http.StatusBadRequest, "name is required")
		return
	}
	if code == "" {
		response.Error(w, http.StatusBadRequest, "code is required")
		return
	}
	item, err := h.repo.Update(r.Context(), id, Permission{Name: name, Code: code, Description: strings.TrimSpace(req.Description)})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "permission not found")
			return
		}
		if errors.Is(err, ErrDuplicateName) {
			response.Error(w, http.StatusConflict, "permission name already exists")
			return
		}
		if errors.Is(err, ErrDuplicateCode) {
			response.Error(w, http.StatusConflict, "permission code already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "data": item})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "permission not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "permission deleted successfully"})
}

func parseID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid resource id")
		return 0, false
	}
	return id, true
}
