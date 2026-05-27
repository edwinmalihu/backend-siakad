package roles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound        = errors.New("role not found")
	ErrDuplicateName   = errors.New("role name already exists")
	ErrDuplicateCode   = errors.New("role code already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string) ([]Role, error) {
	query := `
		SELECT id, name, code, COALESCE(description, ''), created_at, updated_at, deleted_at
		FROM roles
		WHERE deleted_at IS NULL
	`

	var args []any
	if search != "" {
		query += " AND (name LIKE ? OR code LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern)
	}

	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	items := make([]Role, 0)
	for rows.Next() {
		var item Role
		var deletedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.CreatedAt, &item.UpdatedAt, &deletedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		if deletedAt.Valid {
			item.DeletedAt = &deletedAt.Time
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Role, error) {
	const query = `
		SELECT id, name, code, COALESCE(description, ''), created_at, updated_at, deleted_at
		FROM roles WHERE id = ? AND deleted_at IS NULL LIMIT 1
	`
	var item Role
	var deletedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.CreatedAt, &item.UpdatedAt, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get role by id: %w", err)
	}
	if deletedAt.Valid {
		item.DeletedAt = &deletedAt.Time
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Role) (*Role, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO roles (name, code, description) VALUES (?, ?, ?)
	`, item.Name, item.Code, nullableString(item.Description))
	if err != nil {
		if isDuplicateError(err) {
			if strings.Contains(err.Error(), "uk_roles_name") {
				return nil, ErrDuplicateName
			}
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("insert role: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted role id: %w", err)
	}
	return r.GetByID(ctx, uint64(id))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Role) (*Role, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE roles SET name = ?, code = ?, description = ? WHERE id = ? AND deleted_at IS NULL
	`, item.Name, item.Code, nullableString(item.Description), id)
	if err != nil {
		if isDuplicateError(err) {
			if strings.Contains(err.Error(), "uk_roles_name") {
				return nil, ErrDuplicateName
			}
			return nil, ErrDuplicateCode
		}
		return nil, fmt.Errorf("update role: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE roles SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func nullableString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "1062")
}
