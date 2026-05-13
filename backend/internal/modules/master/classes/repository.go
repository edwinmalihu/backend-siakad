package classes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound             = errors.New("class not found")
	ErrAcademicYearNotFound = errors.New("academic year not found")
	ErrDepartmentNotFound   = errors.New("department not found")
	ErrGradeLevelNotFound   = errors.New("grade level not found")
	ErrDuplicateScope       = errors.New("class already exists in the selected academic year, department, and grade level")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, academicYearID, departmentID, gradeLevelID uint64, isActive *bool) ([]Class, error) {
	query := `
		SELECT
			c.id,
			c.academic_year_id,
			ay.name,
			c.department_id,
			d.code,
			d.name,
			c.grade_level_id,
			gl.code,
			gl.name,
			c.name,
			c.is_active,
			c.created_at,
			c.updated_at,
			c.deleted_at
		FROM classes c
		INNER JOIN academic_years ay ON ay.id = c.academic_year_id
		INNER JOIN departments d ON d.id = c.department_id
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id
		WHERE c.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (c.name LIKE ? OR ay.name LIKE ? OR d.code LIKE ? OR d.name LIKE ? OR gl.code LIKE ? OR gl.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if academicYearID > 0 {
		query += " AND c.academic_year_id = ?"
		args = append(args, academicYearID)
	}

	if departmentID > 0 {
		query += " AND c.department_id = ?"
		args = append(args, departmentID)
	}

	if gradeLevelID > 0 {
		query += " AND c.grade_level_id = ?"
		args = append(args, gradeLevelID)
	}

	if isActive != nil {
		query += " AND c.is_active = ?"
		args = append(args, *isActive)
	}

	query += " ORDER BY ay.start_date DESC, gl.sort_order ASC, d.code ASC, c.name ASC, c.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query classes: %w", err)
	}
	defer rows.Close()

	items := make([]Class, 0)
	for rows.Next() {
		var item Class
		if err := rows.Scan(
			&item.ID,
			&item.AcademicYearID,
			&item.AcademicYearName,
			&item.DepartmentID,
			&item.DepartmentCode,
			&item.DepartmentName,
			&item.GradeLevelID,
			&item.GradeLevelCode,
			&item.GradeLevelName,
			&item.Name,
			&item.IsActive,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan class: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate classes: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Class, error) {
	const query = `
		SELECT
			c.id,
			c.academic_year_id,
			ay.name,
			c.department_id,
			d.code,
			d.name,
			c.grade_level_id,
			gl.code,
			gl.name,
			c.name,
			c.is_active,
			c.created_at,
			c.updated_at,
			c.deleted_at
		FROM classes c
		INNER JOIN academic_years ay ON ay.id = c.academic_year_id
		INNER JOIN departments d ON d.id = c.department_id
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id
		WHERE c.id = ?
		  AND c.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
		LIMIT 1
	`

	var item Class
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.DepartmentID,
		&item.DepartmentCode,
		&item.DepartmentName,
		&item.GradeLevelID,
		&item.GradeLevelCode,
		&item.GradeLevelName,
		&item.Name,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get class by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Class) (*Class, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create class transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return nil, err
	}
	if err := ensureDepartmentExists(ctx, tx, item.DepartmentID); err != nil {
		return nil, err
	}
	if err := ensureGradeLevelExists(ctx, tx, item.GradeLevelID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO classes (academic_year_id, department_id, grade_level_id, name, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, item.AcademicYearID, item.DepartmentID, item.GradeLevelID, item.Name, item.IsActive)
	if err != nil {
		if isDuplicateScopeError(err) {
			return nil, ErrDuplicateScope
		}
		return nil, fmt.Errorf("insert class: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted class id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create class transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Class) (*Class, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update class transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return nil, err
	}
	if err := ensureDepartmentExists(ctx, tx, item.DepartmentID); err != nil {
		return nil, err
	}
	if err := ensureGradeLevelExists(ctx, tx, item.GradeLevelID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE classes
		SET academic_year_id = ?, department_id = ?, grade_level_id = ?, name = ?, is_active = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.AcademicYearID, item.DepartmentID, item.GradeLevelID, item.Name, item.IsActive, id)
	if err != nil {
		if isDuplicateScopeError(err) {
			return nil, ErrDuplicateScope
		}
		return nil, fmt.Errorf("update class: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated class affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update class transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE classes
		SET deleted_at = NOW(), is_active = 0
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete class: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted class affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func ensureAcademicYearExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM academic_years WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check academic year existence: %w", err)
	}
	if !exists {
		return ErrAcademicYearNotFound
	}
	return nil
}

func ensureDepartmentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM departments WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check department existence: %w", err)
	}
	if !exists {
		return ErrDepartmentNotFound
	}
	return nil
}

func ensureGradeLevelExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM grade_levels WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check grade level existence: %w", err)
	}
	if !exists {
		return ErrGradeLevelNotFound
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
