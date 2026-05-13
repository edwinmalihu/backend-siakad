package gradelevels

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound      = errors.New("grade level not found")
	ErrDuplicateCode = errors.New("grade level code already exists")
	ErrDuplicateName = errors.New("grade level name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string) ([]GradeLevel, error) {
	query := `
		SELECT id, code, name, sort_order, created_at, updated_at, deleted_at
		FROM grade_levels
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 2)
	if search != "" {
		query += " AND (code LIKE ? OR name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern)
	}

	query += " ORDER BY sort_order ASC, code ASC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query grade levels: %w", err)
	}
	defer rows.Close()

	items := make([]GradeLevel, 0)
	for rows.Next() {
		var item GradeLevel
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan grade level: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate grade levels: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*GradeLevel, error) {
	const query = `
		SELECT id, code, name, sort_order, created_at, updated_at, deleted_at
		FROM grade_levels
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`

	var item GradeLevel
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.SortOrder,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get grade level by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item GradeLevel) (*GradeLevel, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO grade_levels (code, name, sort_order)
		VALUES (?, ?, ?)
	`, item.Code, item.Name, item.SortOrder)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert grade level: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted grade level id: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item GradeLevel) (*GradeLevel, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE grade_levels
		SET code = ?, name = ?, sort_order = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Code, item.Name, item.SortOrder, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update grade level: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated grade level affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE grade_levels
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete grade level: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted grade level affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}

	message := strings.ToLower(mysqlErr.Message)
	switch {
	case strings.Contains(message, "uk_grade_levels_active_code"), strings.Contains(message, "active_code"):
		return ErrDuplicateCode
	case strings.Contains(message, "uk_grade_levels_active_name"), strings.Contains(message, "active_name"):
		return ErrDuplicateName
	default:
		return fmt.Errorf("duplicate grade level data: %w", err)
	}
}
