package disciplinecategories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound      = errors.New("discipline category not found")
	ErrDuplicateName = errors.New("discipline category name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string) ([]DisciplineCategory, error) {
	query := `
		SELECT id, name, point, description, created_at, updated_at, deleted_at
		FROM discipline_categories
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 2)
	if search != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern)
	}

	query += " ORDER BY point DESC, name ASC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query discipline categories: %w", err)
	}
	defer rows.Close()

	items := make([]DisciplineCategory, 0)
	for rows.Next() {
		item, err := scanDisciplineCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discipline categories: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*DisciplineCategory, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, point, description, created_at, updated_at, deleted_at
		FROM discipline_categories
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, id)

	item, err := scanDisciplineCategory(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get discipline category by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item DisciplineCategory) (*DisciplineCategory, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO discipline_categories (name, point, description)
		VALUES (?, ?, ?)
	`, item.Name, item.Point, nullableString(item.Description))
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert discipline category: %w", err)
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted discipline category id: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item DisciplineCategory) (*DisciplineCategory, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE discipline_categories
		SET name = ?, point = ?, description = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Name, item.Point, nullableString(item.Description), id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update discipline category: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated discipline category affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE discipline_categories
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete discipline category: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted discipline category affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDisciplineCategory(s scanner) (DisciplineCategory, error) {
	var item DisciplineCategory
	var description sql.NullString
	err := s.Scan(
		&item.ID,
		&item.Name,
		&item.Point,
		&description,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return DisciplineCategory{}, err
	}
	if description.Valid {
		item.Description = description.String
	}
	return item, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}
	message := strings.ToLower(mysqlErr.Message)
	if strings.Contains(message, "uk_discipline_categories_active_name") {
		return ErrDuplicateName
	}
	return fmt.Errorf("duplicate discipline category data: %w", err)
}
