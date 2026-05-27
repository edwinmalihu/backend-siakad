package auditlogs

import "time"

type AuditLog struct {
	ID          uint64  `json:"id"`
	UserID      *uint64 `json:"user_id,omitempty"`
	UserName    string  `json:"user_name,omitempty"`
	Module      string  `json:"module"`
	Action      string  `json:"action"`
	EntityType  string  `json:"entity_type,omitempty"`
	EntityID    *uint64 `json:"entity_id,omitempty"`
	PayloadJSON string  `json:"payload_json,omitempty"`
	IPAddress   string  `json:"ip_address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
