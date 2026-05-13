package subjects

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound           = errors.New("subject not found")
	ErrDepartmentNotFound = errors.New("department not found")
	ErrGradeLevelNotFound = errors.New("grade level not found")
	ErrDuplicateCode      = errors.New("subject code already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, departmentID, gradeLevelID uint64) ([]Subject, error) {
	query := `
		SELECT
			s.id,
			s.department_id,
			d.code,
			d.name,
			s.grade_level_id,
			gl.code,
			gl.name,
			s.code,
			s.name,
			s.subject_type,
			s.kkm,
			s.created_at,
			s.updated_at,
			s.deleted_at
		FROM subjects s
		INNER JOIN departments d ON d.id = s.department_id
		INNER JOIN grade_levels gl ON gl.id = s.grade_level_id
		WHERE s.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (s.code LIKE ? OR s.name LIKE ? OR s.subject_type LIKE ? OR d.code LIKE ? OR d.name LIKE ? OR gl.code LIKE ? OR gl.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if departmentID > 0 {
		query += " AND s.department_id = ?"
		args = append(args, departmentID)
	}

	if gradeLevelID > 0 {
		query += " AND s.grade_level_id = ?"
		args = append(args, gradeLevelID)
	}

	query += " ORDER BY d.code ASC, gl.sort_order ASC, s.code ASC, s.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query subjects: %w", err)
	}
	defer rows.Close()

	items := make([]Subject, 0)
	for rows.Next() {
		var item Subject
		var subjectType sql.NullString
		var kkm sql.NullFloat64
		if err := rows.Scan(
			&item.ID,
			&item.DepartmentID,
			&item.DepartmentCode,
			&item.DepartmentName,
			&item.GradeLevelID,
			&item.GradeLevelCode,
			&item.GradeLevelName,
			&item.Code,
			&item.Name,
			&subjectType,
			&kkm,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}

		if subjectType.Valid {
			item.SubjectType = subjectType.String
		}
		if kkm.Valid {
			item.KKM = &kkm.Float64
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Subject, error) {
	const query = `
		SELECT
			s.id,
			s.department_id,
			d.code,
			d.name,
			s.grade_level_id,
			gl.code,
			gl.name,
			s.code,
			s.name,
			s.subject_type,
			s.kkm,
			s.created_at,
			s.updated_at,
			s.deleted_at
		FROM subjects s
		INNER JOIN departments d ON d.id = s.department_id
		INNER JOIN grade_levels gl ON gl.id = s.grade_level_id
		WHERE s.id = ?
		  AND s.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
		LIMIT 1
	`

	var item Subject
	var subjectType sql.NullString
	var kkm sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.DepartmentID,
		&item.DepartmentCode,
		&item.DepartmentName,
		&item.GradeLevelID,
		&item.GradeLevelCode,
		&item.GradeLevelName,
		&item.Code,
		&item.Name,
		&subjectType,
		&kkm,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get subject by id: %w", err)
	}

	if subjectType.Valid {
		item.SubjectType = subjectType.String
	}
	if kkm.Valid {
		item.KKM = &kkm.Float64
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Subject) (*Subject, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create subject transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureDepartmentExists(ctx, tx, item.DepartmentID); err != nil {
		return nil, err
	}
	if err := ensureGradeLevelExists(ctx, tx, item.GradeLevelID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO subjects (department_id, grade_level_id, code, name, subject_type, kkm)
		VALUES (?, ?, ?, ?, ?, ?)
	`, item.DepartmentID, item.GradeLevelID, item.Code, item.Name, nullableString(item.SubjectType), item.KKM)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert subject: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted subject id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create subject transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Subject) (*Subject, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update subject transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureDepartmentExists(ctx, tx, item.DepartmentID); err != nil {
		return nil, err
	}
	if err := ensureGradeLevelExists(ctx, tx, item.GradeLevelID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE subjects
		SET department_id = ?, grade_level_id = ?, code = ?, name = ?, subject_type = ?, kkm = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.DepartmentID, item.GradeLevelID, item.Code, item.Name, nullableString(item.SubjectType), item.KKM, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update subject: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated subject affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update subject transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE subjects
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete subject: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted subject affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
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
	case strings.Contains(message, "uk_subjects_active_code"), strings.Contains(message, "active_code"):
		return ErrDuplicateCode
	default:
		return fmt.Errorf("duplicate subject data: %w", err)
	}
}
