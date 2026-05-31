package license

import "time"

type InstalledLicense struct {
	ID                int64      `json:"id"`
	LicenseKey        string     `json:"license_key"`
	Tier              string     `json:"tier"`
	DeviceFingerprint string     `json:"device_fingerprint"`
	StartsAt          time.Time  `json:"starts_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	TrialCount        int        `json:"trial_count"`
	ClientName        string     `json:"client_name"`
	LastAlertAt       *time.Time `json:"last_alert_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type LicenseStatus struct {
	HasLicense    bool             `json:"has_license"`
	IsActive      bool             `json:"is_active"`
	IsExpired     bool             `json:"is_expired"`
	IsExpiringSoon bool            `json:"is_expiring_soon"`
	DaysRemaining int              `json:"days_remaining"`
	License       *InstalledLicense `json:"license,omitempty"`
}

type ActivateRequest struct {
	LicenseKey string `json:"license_key"`
	ClientName string `json:"client_name"`
}

type TrialRequest struct {
	ClientName string `json:"client_name"`
}

// Responses from License Generator API
type GeneratorValidateResponse struct {
	Valid      bool   `json:"valid"`
	Tier       string `json:"tier"`
	ExpiresAt  string `json:"expires_at"`
	ClientName string `json:"client_name"`
	Message    string `json:"message"`
}

type GeneratorActivateResponse struct {
	Success   bool   `json:"success"`
	Tier      string `json:"tier"`
	ExpiresAt string `json:"expires_at"`
	Message   string `json:"message"`
}

type GeneratorTrialResponse struct {
	Success     bool   `json:"success"`
	LicenseKey  string `json:"license_key"`
	Tier        string `json:"tier"`
	ExpiresAt   string `json:"expires_at"`
	TrialCount  int    `json:"trial_count"`
	Message     string `json:"message"`
}

type GeneratorStatusResponse struct {
	Key                string `json:"key"`
	Tier               string `json:"tier"`
	ClientName         string `json:"client_name"`
	DeviceFingerprint  string `json:"device_fingerprint"`
	IsUsed             bool   `json:"is_used"`
	StartsAt           string `json:"starts_at"`
	ExpiresAt          string `json:"expires_at"`
	Status             string `json:"status"`
}
