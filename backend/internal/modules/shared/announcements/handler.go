package announcements

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
	mux.HandleFunc("GET /api/v1/shared/announcements", h.List)
	mux.HandleFunc("POST /api/v1/shared/announcements", h.Create)
	mux.HandleFunc("GET /api/v1/shared/announcements/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/shared/announcements/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/shared/announcements/{id}", h.Delete)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	targetScope := strings.TrimSpace(r.URL.Query().Get("target_scope"))
	isPublished, hasPublished, ok := parseOptionalBool(w, r.URL.Query().Get("is_published"), "is_published")
	if !ok {
		return
	}
	var publishedFilter *bool
	if hasPublished {
		publishedFilter = &isPublished
	}
	items, err := h.repo.List(r.Context(), search, targetScope, publishedFilter)
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
			response.Error(w, http.StatusNotFound, "announcement not found")
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
	item, err := buildAnnouncement(req.Title, req.Content, req.TargetScope, req.PublishStart, req.PublishEnd, req.IsPublished)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.repo.Create(r.Context(), item)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
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
	item, err := buildAnnouncement(req.Title, req.Content, req.TargetScope, req.PublishStart, req.PublishEnd, req.IsPublished)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	updated, err := h.repo.Update(r.Context(), id, item)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.Error(w, http.StatusNotFound, "announcement not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
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
			response.Error(w, http.StatusNotFound, "announcement not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"success": true, "message": "announcement deleted successfully"})
}

func parseID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid resource id")
		return 0, false
	}
	return id, true
}

func parseOptionalBool(w http.ResponseWriter, raw, field string) (bool, bool, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, false, true
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		response.Error(w, http.StatusBadRequest, fmt.Sprintf("%s must be true or false", field))
		return false, false, false
	}
	return parsed, true, true
}

func buildAnnouncement(title, content, targetScope, publishStartRaw, publishEndRaw string, isPublished bool) (Announcement, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	targetScope = strings.TrimSpace(targetScope)
	if title == "" {
		return Announcement{}, fmt.Errorf("title is required")
	}
	if content == "" {
		return Announcement{}, fmt.Errorf("content is required")
	}

	publishStart, err := parseOptionalDateTime(publishStartRaw)
	if err != nil {
		return Announcement{}, fmt.Errorf("publish_start must be a valid datetime")
	}
	publishEnd, err := parseOptionalDateTime(publishEndRaw)
	if err != nil {
		return Announcement{}, fmt.Errorf("publish_end must be a valid datetime")
	}
	if publishStart != nil && publishEnd != nil && publishEnd.Before(*publishStart) {
		return Announcement{}, fmt.Errorf("publish_end must be after publish_start")
	}

	return Announcement{
		Title:        title,
		Content:      content,
		TargetScope:  targetScope,
		PublishStart: publishStart,
		PublishEnd:   publishEnd,
		IsPublished:  isPublished,
	}, nil
}

func parseOptionalDateTime(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			value := parsed
			return &value, nil
		}
	}
	return nil, fmt.Errorf("invalid datetime")
}
