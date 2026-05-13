package rooms

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound      = errors.New("room not found")
	ErrDuplicateCode = errors.New("room code already exists")
	ErrDuplicateName = errors.New("room name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string) ([]Room, error) {
	query := `
		SELECT id, code, name, type, capacity, created_at, updated_at, deleted_at
		FROM rooms
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 3)
	if search != "" {
		query += " AND (code LIKE ? OR name LIKE ? OR type LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	query += " ORDER BY code ASC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query rooms: %w", err)
	}
	defer rows.Close()

	items := make([]Room, 0)
	for rows.Next() {
		var item Room
		var roomType sql.NullString
		var capacity sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&roomType,
			&capacity,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		if roomType.Valid {
			item.Type = roomType.String
		}
		if capacity.Valid {
			value := int(capacity.Int64)
			item.Capacity = &value
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate rooms: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Room, error) {
	const query = `
		SELECT id, code, name, type, capacity, created_at, updated_at, deleted_at
		FROM rooms
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`

	var item Room
	var roomType sql.NullString
	var capacity sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&roomType,
		&capacity,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get room by id: %w", err)
	}

	if roomType.Valid {
		item.Type = roomType.String
	}
	if capacity.Valid {
		value := int(capacity.Int64)
		item.Capacity = &value
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Room) (*Room, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO rooms (code, name, type, capacity)
		VALUES (?, ?, ?, ?)
	`, item.Code, item.Name, nullableString(item.Type), item.Capacity)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert room: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted room id: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Room) (*Room, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE rooms
		SET code = ?, name = ?, type = ?, capacity = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Code, item.Name, nullableString(item.Type), item.Capacity, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update room: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated room affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE rooms
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete room: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted room affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func nullableString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}

	message := strings.ToLower(mysqlErr.Message)
	switch {
	case strings.Contains(message, "uk_rooms_active_code"), strings.Contains(message, "active_code"):
		return ErrDuplicateCode
	case strings.Contains(message, "uk_rooms_active_name"), strings.Contains(message, "active_name"):
		return ErrDuplicateName
	default:
		return fmt.Errorf("duplicate room data: %w", err)
	}
}
