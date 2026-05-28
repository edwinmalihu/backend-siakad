package auditlogs

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// LogAudit is a shared helper to write an audit log entry from any HTTP handler.
// The caller should pass the userID extracted from auth context to avoid import cycles.
func LogAudit(ctx context.Context, r *http.Request, repo *Repository, module, action, entityType string, entityID *uint64, userID *uint64, payload interface{}) {
	if repo == nil {
		return
	}

	// Marshal payload to JSON
	payloadJSON := ""
	if payload != nil {
		if bytes, err := json.Marshal(payload); err == nil {
			payloadJSON = string(bytes)
		}
	}

	// Capture IP address
	ipAddress := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			ipAddress = strings.TrimSpace(parts[0])
		}
	}

	auditLog := &AuditLog{
		UserID:      userID,
		Module:      module,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		PayloadJSON: payloadJSON,
		IPAddress:   ipAddress,
	}

	_ = repo.Create(ctx, auditLog)
}

// LogAuditWithID is a convenience wrapper that takes a uint64 entityID.
func LogAuditWithID(ctx context.Context, r *http.Request, repo *Repository, module, action, entityType string, entityID uint64, userID *uint64, payload interface{}) {
	LogAudit(ctx, r, repo, module, action, entityType, &entityID, userID, payload)
}
