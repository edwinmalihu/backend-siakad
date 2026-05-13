package semesters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound             = errors.New("semester not found")
	ErrAcademicYearNotFound = errors.New("academic year not found")
	ErrDuplicateScope       = errors.New("semester code already exists in the selected academic year")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, academicYearID uint64, isActive *bool) ([]Semester, error) {
	query := `
		SELECT
			s.id,
			s.academic_year_id,
			ay.name,
			s.name,
			s.code,
			s.is_active,
			s.created_at,
			s.updated_at,
			s.deleted_at
		FROM semesters s
		INNER JOIN academic_years ay ON ay.id = s.academic_year_id
		WHERE s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
	`

	args := make([]any, 0, 3)
	if search != "" {
		query += " AND (s.name LIKE ? OR s.code LIKE ? OR ay.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern)
	}

	if academicYearID > 0 {
		query += " AND s.academic_year_id = ?"
		args = append(args, academicYearID)
	}

	if isActive != nil {
		query += " AND s.is_active = ?"
		args = append(args, *isActive)
	}

	query += " ORDER BY ay.start_date DESC, s.code ASC, s.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query semesters: %w", err)
	}
	defer rows.Close()

	items := make([]Semester, 0)
	for rows.Next() {
		var item Semester
		if err := rows.Scan(
			&item.ID,
			&item.AcademicYearID,
			&item.AcademicYear,
			&item.Name,
			&item.Code,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan semester: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate semesters: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Semester, error) {
	const query = `
		SELECT
			s.id,
			s.academic_year_id,
			ay.name,
			s.name,
			s.code,
			s.is_active,
			s.created_at,
			s.updated_at,
			s.deleted_at
		FROM semesters s
		INNER JOIN academic_years ay ON ay.id = s.academic_year_id
		WHERE s.id = ?
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		LIMIT 1
	`

	var item Semester
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.AcademicYearID,
		&item.AcademicYear,
		&item.Name,
		&item.Code,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get semester by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Semester) (*Semester, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create semester transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return nil, err
	}

	if item.IsActive {
		if _, err := tx.ExecContext(ctx, `UPDATE semesters SET is_active = 0 WHERE deleted_at IS NULL`); err != nil {
			return nil, fmt.Errorf("reset active semesters: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO semesters (academic_year_id, name, code, is_active)
		VALUES (?, ?, ?, ?)
	`, item.AcademicYearID, item.Name, item.Code, item.IsActive)
	if err != nil {
		if isDuplicateScopeError(err) {
			return nil, ErrDuplicateScope
		}
		return nil, fmt.Errorf("insert semester: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted semester id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create semester transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Semester) (*Semester, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update semester transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return nil, err
	}

	if item.IsActive {
		if _, err := tx.ExecContext(ctx, `UPDATE semesters SET is_active = 0 WHERE deleted_at IS NULL AND id <> ?`, id); err != nil {
			return nil, fmt.Errorf("reset active semesters: %w", err)
		}
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE semesters
		SET academic_year_id = ?, name = ?, code = ?, is_active = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.AcademicYearID, item.Name, item.Code, item.IsActive, id)
	if err != nil {
		if isDuplicateScopeError(err) {
			return nil, ErrDuplicateScope
		}
		return nil, fmt.Errorf("update semester: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated semester affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update semester transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE semesters
		SET deleted_at = NOW(), is_active = 0
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete semester: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted semester affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func ensureAcademicYearExists(ctx context.Context, tx *sql.Tx, academicYearID uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM academic_years
			WHERE id = ? AND deleted_at IS NULL
		)
	`, academicYearID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check academic year existence: %w", err)
	}
	if !exists {
		return ErrAcademicYearNotFound
	}

	return nil
}

func isDuplicateScopeError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}

	return mysqlErr.Number == 1062
}
