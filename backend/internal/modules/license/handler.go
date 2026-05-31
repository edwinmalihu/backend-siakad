package license

import (
	"encoding/json"
	"net/http"
	"time"

	"siakad/backend/internal/modules/auth"
	"siakad/backend/internal/response"
)

type Handler struct {
	repo   *Repository
	client *GeneratorClient
}

func NewHandler(repo *Repository, client *GeneratorClient) *Handler {
	return &Handler{repo: repo, client: client}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/license/status", h.Status)
	mux.HandleFunc("POST /api/v1/license/activate", h.Activate)
	mux.HandleFunc("POST /api/v1/license/trial", h.Trial)
	mux.HandleFunc("POST /api/v1/license/validate", h.ValidateRemotely)
	mux.HandleFunc("GET /api/v1/license/info", h.LicenseInfo)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	license, err := h.repo.GetActive(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to check license")
		return
	}

	if license == nil {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": LicenseStatus{
				HasLicense: false,
			},
		})
		return
	}

	now := time.Now()
	isExpired := now.After(license.ExpiresAt)
	daysRemaining := int(time.Until(license.ExpiresAt).Hours() / 24)
	isExpiringSoon := !isExpired && daysRemaining <= 30

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": LicenseStatus{
			HasLicense:     true,
			IsActive:       !isExpired,
			IsExpired:      isExpired,
			IsExpiringSoon: isExpiringSoon,
			DaysRemaining:  daysRemaining,
			License:        license,
		},
	})
}

func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	var req ActivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.LicenseKey == "" {
		response.Error(w, http.StatusBadRequest, "license_key is required")
		return
	}

	user := auth.GetUserFromContext(r.Context())
	deviceFingerprint := getDeviceFingerprint(r)
	clientName := req.ClientName
	if clientName == "" && user != nil {
		clientName = user.Username
	}

	// Activate with Generator API
	actResp, err := h.client.Activate(r.Context(), req.LicenseKey, deviceFingerprint, clientName)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Aktivasi gagal: "+err.Error())
		return
	}

	// Parse expires_at
	expiresAt, _ := time.Parse(time.RFC3339, actResp.ExpiresAt)
	startsAt := time.Now()

	// Save locally
	license := &InstalledLicense{
		LicenseKey:        req.LicenseKey,
		Tier:              actResp.Tier,
		DeviceFingerprint: deviceFingerprint,
		StartsAt:          startsAt,
		ExpiresAt:         expiresAt,
		ClientName:        clientName,
	}

	if err := h.repo.Save(r.Context(), license); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to save license locally")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "License activated successfully",
		"data": map[string]any{
			"tier":       actResp.Tier,
			"expires_at": actResp.ExpiresAt,
		},
	})
}

func (h *Handler) Trial(w http.ResponseWriter, r *http.Request) {
	var req TrialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	deviceFingerprint := getDeviceFingerprint(r)
	clientName := req.ClientName

	// Get current trial count
	trialCount, _ := h.repo.GetTrialCount(r.Context(), deviceFingerprint)

	// Start trial via Generator API
	trialResp, err := h.client.Trial(r.Context(), deviceFingerprint, clientName, trialCount)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Trial gagal: "+err.Error())
		return
	}

	// Parse expires_at
	expiresAt, _ := time.Parse(time.RFC3339, trialResp.ExpiresAt)
	startsAt := time.Now()

	// Save locally
	license := &InstalledLicense{
		LicenseKey:        trialResp.LicenseKey,
		Tier:              "trial",
		DeviceFingerprint: deviceFingerprint,
		StartsAt:          startsAt,
		ExpiresAt:         expiresAt,
		TrialCount:        trialResp.TrialCount,
		ClientName:        clientName,
	}

	if err := h.repo.Save(r.Context(), license); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to save trial license locally")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Trial started successfully",
		"data": map[string]any{
			"license_key":  trialResp.LicenseKey,
			"tier":         "trial",
			"expires_at":   trialResp.ExpiresAt,
			"trial_count":  trialResp.TrialCount,
		},
	})
}

func (h *Handler) ValidateRemotely(w http.ResponseWriter, r *http.Request) {
	license, err := h.repo.GetActive(r.Context())
	if err != nil || license == nil {
		response.Error(w, http.StatusBadRequest, "no license installed")
		return
	}

	deviceFingerprint := license.DeviceFingerprint

	valResp, err := h.client.Validate(r.Context(), license.LicenseKey, deviceFingerprint)
	if err != nil {
		response.Error(w, http.StatusBadGateway, "failed to reach license generator: "+err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    valResp,
	})
}

func (h *Handler) LicenseInfo(w http.ResponseWriter, r *http.Request) {
	license, err := h.repo.GetActive(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get license info")
		return
	}

	if license == nil {
		response.JSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    nil,
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    license,
	})
}

func getDeviceFingerprint(r *http.Request) string {
	// Generate a simple fingerprint from request headers
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return host
}
