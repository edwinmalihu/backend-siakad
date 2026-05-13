package departments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound      = errors.New("department not found")
	ErrDuplicateCode = errors.New("department code already exists")
	ErrDuplicateName = errors.New("department name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string) ([]Department, error) {
	query := `
		SELECT id, code, name, program_name, field_name, description, created_at, updated_at, deleted_at
		FROM departments
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 4)
	if search != "" {
		query += " AND (code LIKE ? OR name LIKE ? OR program_name LIKE ? OR field_name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	query += " ORDER BY code ASC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query departments: %w", err)
	}
	defer rows.Close()

	items := make([]Department, 0)
	for rows.Next() {
		var item Department
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.ProgramName,
			&item.FieldName,
			&item.Description,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan department: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate departments: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Department, error) {
	const query = `
		SELECT id, code, name, program_name, field_name, description, created_at, updated_at, deleted_at
		FROM departments
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`

	var item Department
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.ProgramName,
		&item.FieldName,
		&item.Description,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get department by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Department) (*Department, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO departments (code, name, program_name, field_name, description)
		VALUES (?, ?, ?, ?, ?)
	`, item.Code, item.Name, nullableString(item.ProgramName), nullableString(item.FieldName), nullableString(item.Description))
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert department: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted department id: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Department) (*Department, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE departments
		SET code = ?, name = ?, program_name = ?, field_name = ?, description = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Code, item.Name, nullableString(item.ProgramName), nullableString(item.FieldName), nullableString(item.Description), id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update department: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated department affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE departments
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete department: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted department affected rows: %w", err)
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
	case strings.Contains(message, "uk_departments_active_code"), strings.Contains(message, "active_code"):
		return ErrDuplicateCode
	case strings.Contains(message, "uk_departments_active_name"), strings.Contains(message, "active_name"):
		return ErrDuplicateName
	default:
		return fmt.Errorf("duplicate department data: %w", err)
	}
}
