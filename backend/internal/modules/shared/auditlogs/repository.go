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

func (r *Repository) Create(ctx context.Context, log *AuditLog) error {
	const query = `
		INSERT INTO audit_logs (user_id, module, action, entity_type, entity_id, payload_json, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	var entityType, payloadJSON, ipAddress interface{}
	if log.EntityType != "" {
		entityType = log.EntityType
	}
	if log.PayloadJSON != "" {
		payloadJSON = log.PayloadJSON
	}
	if log.IPAddress != "" {
		ipAddress = log.IPAddress
	}

	result, err := r.db.ExecContext(ctx, query,
		log.UserID,
		log.Module,
		log.Action,
		entityType,
		log.EntityID,
		payloadJSON,
		ipAddress,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	log.ID = uint64(id)

	return nil
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
			al.created_at,
			al.logout_time
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
			&item.LogoutTime,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		if userID.Valid {
			uid := uint64(userID.Int64)
			item.UserID = &uid
		}
		// For login records, login_time = created_at
		if item.Action == "login" {
			item.LoginTime = &item.CreatedAt
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
			al.created_at,
			al.logout_time
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
		&item.LogoutTime,
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

	// For login records, login_time = created_at
	item.LoginTime = &item.CreatedAt

	return &item, nil
}

// FindLatestLogin finds the most recent login audit log for a user that has no logout_time yet.
func (r *Repository) FindLatestLogin(ctx context.Context, userID uint64) (*AuditLog, error) {
	const query = `
		SELECT id, created_at
		FROM audit_logs
		WHERE user_id = ?
		  AND action = 'login'
		  AND logout_time IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var item AuditLog
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&item.ID, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find latest login: %w", err)
	}

	return &item, nil
}

// UpdateLogoutTime sets the logout_time on an audit log record.
func (r *Repository) UpdateLogoutTime(ctx context.Context, id uint64) error {
	const query = `
		UPDATE audit_logs
		SET logout_time = NOW()
		WHERE id = ? AND logout_time IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("update logout time: %w", err)
	}

	return nil
}
