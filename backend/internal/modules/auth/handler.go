package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"siakad/backend/internal/response"
)

type Handler struct {
	repo    *Repository
	service *Service
}

func NewHandler(repo *Repository, service *Service) *Handler {
	return &Handler{
		repo:    repo,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("GET /api/v1/auth/me", h.Me)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	identifier := normalizeIdentifier(req.Identifier)
	if identifier == "" {
		identifier = normalizeIdentifier(req.Username)
	}
	if identifier == "" {
		response.Error(w, http.StatusBadRequest, "identifier is required")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		response.Error(w, http.StatusBadRequest, "password is required")
		return
	}

	user, err := h.repo.FindByIdentifier(r.Context(), identifier)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !user.IsActive {
		response.Error(w, http.StatusForbidden, "user account is inactive")
		return
	}

	if err := h.service.VerifyPassword(req.Password, user.PasswordHash); err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	if err := h.repo.UpdateLastLogin(r.Context(), user.ID); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, expiresAt, err := h.service.GenerateToken(*user)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	user.PasswordHash = ""

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": LoginResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresAt:   expiresAt,
			User:        *user,
		},
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	token, ok := extractBearerToken(r.Header.Get("Authorization"))
	if !ok {
		response.Error(w, http.StatusUnauthorized, "authorization token is required")
		return
	}

	claims, err := h.service.ParseToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	user, err := h.repo.FindByID(r.Context(), claims.Sub)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Error(w, http.StatusUnauthorized, "user account is not available")
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !user.IsActive {
		response.Error(w, http.StatusForbidden, "user account is inactive")
		return
	}

	user.PasswordHash = ""

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    user,
	})
}

func extractBearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

func normalizeIdentifier(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
