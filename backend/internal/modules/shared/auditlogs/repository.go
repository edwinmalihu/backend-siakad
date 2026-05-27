package auditlogs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("audit log not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, module, action string, userID uint64) ([]AuditLog, error) {
	query := `
		SELECT
			al.id,
			al.user_id,
			COALESCE(u.username, '') AS user_name,
			al.module,
			al.action,
			COALESCE(al.entity_type, '') AS entity_type,
			al.entity_id,
			COALESCE(al.payload_json, '') AS payload_json,
			COALESCE(al.ip_address, '') AS ip_address,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.user_id
		WHERE 1=1
	`

	args := make([]any, 0, 4)

	if search != "" {
		query += " AND (u.full_name LIKE ? OR al.module LIKE ? OR al.action LIKE ? OR al.entity_type LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	if module != "" {
		query += " AND al.module = ?"
		args = append(args, module)
	}

	if action != "" {
		query += " AND al.action = ?"
		args = append(args, action)
	}

	if userID > 0 {
		query += " AND al.user_id = ?"
		args = append(args, userID)
	}

	query += " ORDER BY al.created_at DESC, al.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	items := make([]AuditLog, 0)
	for rows.Next() {
		var item AuditLog
		var userID sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&userID,
			&item.UserName,
			&item.Module,
			&item.Action,
			&item.EntityType,
			&item.EntityID,
			&item.PayloadJSON,
			&item.IPAddress,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		if userID.Valid {
			uid := uint64(userID.Int64)
			item.UserID = &uid
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*AuditLog, error) {
	const query = `
		SELECT
			al.id,
			al.user_id,
			COALESCE(u.username, '') AS user_name,
			al.module,
			al.action,
			COALESCE(al.entity_type, '') AS entity_type,
			al.entity_id,
			COALESCE(al.payload_json, '') AS payload_json,
			COALESCE(al.ip_address, '') AS ip_address,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.user_id
		WHERE al.id = ?
		LIMIT 1
	`

	var item AuditLog
	var userID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&userID,
		&item.UserName,
		&item.Module,
		&item.Action,
		&item.EntityType,
		&item.EntityID,
		&item.PayloadJSON,
		&item.IPAddress,
		&item.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get audit log by id: %w", err)
	}
	if userID.Valid {
		uid := uint64(userID.Int64)
		item.UserID = &uid
	}

	return &item, nil
}
