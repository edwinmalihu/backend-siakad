package academicyears

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound      = errors.New("academic year not found")
	ErrDuplicateName = errors.New("academic year name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, isActive *bool) ([]AcademicYear, error) {
	query := `
		SELECT id, name, start_date, end_date, is_active, created_at, updated_at, deleted_at
		FROM academic_years
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 2)
	if search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+strings.TrimSpace(search)+"%")
	}

	if isActive != nil {
		query += " AND is_active = ?"
		args = append(args, *isActive)
	}

	query += " ORDER BY start_date DESC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query academic years: %w", err)
	}
	defer rows.Close()

	items := make([]AcademicYear, 0)
	for rows.Next() {
		var item AcademicYear
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.StartDate,
			&item.EndDate,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan academic year: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate academic years: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*AcademicYear, error) {
	const query = `
		SELECT id, name, start_date, end_date, is_active, created_at, updated_at, deleted_at
		FROM academic_years
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`

	var item AcademicYear
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.StartDate,
		&item.EndDate,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get academic year by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item AcademicYear) (*AcademicYear, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create academic year transaction: %w", err)
	}
	defer tx.Rollback()

	if item.IsActive {
		if _, err := tx.ExecContext(ctx, `UPDATE academic_years SET is_active = 0 WHERE deleted_at IS NULL`); err != nil {
			return nil, fmt.Errorf("reset active academic years: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO academic_years (name, start_date, end_date, is_active)
		VALUES (?, ?, ?, ?)
	`, item.Name, item.StartDate, item.EndDate, item.IsActive)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert academic year: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted academic year id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create academic year transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item AcademicYear) (*AcademicYear, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update academic year transaction: %w", err)
	}
	defer tx.Rollback()

	if item.IsActive {
		if _, err := tx.ExecContext(ctx, `UPDATE academic_years SET is_active = 0 WHERE deleted_at IS NULL AND id <> ?`, id); err != nil {
			return nil, fmt.Errorf("reset active academic years: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE academic_years
		SET name = ?, start_date = ?, end_date = ?, is_active = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Name, item.StartDate, item.EndDate, item.IsActive, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update academic year: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated academic year affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update academic year transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE academic_years
		SET deleted_at = NOW(), is_active = 0
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete academic year: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted academic year affected rows: %w", err)
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
	case strings.Contains(message, "uk_academic_years_active_name"), strings.Contains(message, "active_name"):
		return ErrDuplicateName
	default:
		return fmt.Errorf("duplicate academic year data: %w", err)
	}
}
