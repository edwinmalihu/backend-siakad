package extracurricularmembers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound                = errors.New("extracurricular member not found")
	ErrExtracurricularNotFound = errors.New("extracurricular not found")
	ErrStudentNotFound         = errors.New("student not found")
	ErrAcademicYearNotFound    = errors.New("academic year not found")
	ErrDuplicateScope          = errors.New("student membership already exists for the selected extracurricular and academic year")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, status string, extracurricularID, studentID, academicYearID uint64) ([]ExtracurricularMember, error) {
	query := `
		SELECT
			em.id,
			em.extracurricular_id,
			e.name,
			em.student_id,
			s.nis,
			s.full_name,
			em.academic_year_id,
			ay.name,
			em.status,
			em.created_at,
			em.updated_at,
			em.deleted_at
		FROM extracurricular_members em
		INNER JOIN extracurriculars e ON e.id = em.extracurricular_id
		INNER JOIN students s ON s.id = em.student_id
		INNER JOIN academic_years ay ON ay.id = em.academic_year_id
		WHERE em.deleted_at IS NULL
		  AND e.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (e.name LIKE ? OR s.nis LIKE ? OR s.full_name LIKE ? OR ay.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if status != "" {
		query += " AND em.status = ?"
		args = append(args, status)
	}
	if extracurricularID > 0 {
		query += " AND em.extracurricular_id = ?"
		args = append(args, extracurricularID)
	}
	if studentID > 0 {
		query += " AND em.student_id = ?"
		args = append(args, studentID)
	}
	if academicYearID > 0 {
		query += " AND em.academic_year_id = ?"
		args = append(args, academicYearID)
	}

	query += " ORDER BY ay.start_date DESC, e.name ASC, s.full_name ASC, em.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query extracurricular members: %w", err)
	}
	defer rows.Close()

	items := make([]ExtracurricularMember, 0)
	for rows.Next() {
		item, err := scanExtracurricularMember(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate extracurricular members: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*ExtracurricularMember, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			em.id,
			em.extracurricular_id,
			e.name,
			em.student_id,
			s.nis,
			s.full_name,
			em.academic_year_id,
			ay.name,
			em.status,
			em.created_at,
			em.updated_at,
			em.deleted_at
		FROM extracurricular_members em
		INNER JOIN extracurriculars e ON e.id = em.extracurricular_id
		INNER JOIN students s ON s.id = em.student_id
		INNER JOIN academic_years ay ON ay.id = em.academic_year_id
		WHERE em.id = ?
		  AND em.deleted_at IS NULL
		  AND e.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		LIMIT 1
	`, id)

	item, err := scanExtracurricularMember(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get extracurricular member by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item ExtracurricularMember) (*ExtracurricularMember, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create extracurricular member transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO extracurricular_members (extracurricular_id, student_id, academic_year_id, status)
		VALUES (?, ?, ?, ?)
	`, item.ExtracurricularID, item.StudentID, item.AcademicYearID, item.Status)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert extracurricular member: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted extracurricular member id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create extracurricular member transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item ExtracurricularMember) (*ExtracurricularMember, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update extracurricular member transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureMemberExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE extracurricular_members
		SET extracurricular_id = ?, student_id = ?, academic_year_id = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.ExtracurricularID, item.StudentID, item.AcademicYearID, item.Status, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update extracurricular member: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated extracurricular member affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update extracurricular member transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE extracurricular_members
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete extracurricular member: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted extracurricular member affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExtracurricularMember(s scanner) (ExtracurricularMember, error) {
	var item ExtracurricularMember
	err := s.Scan(
		&item.ID,
		&item.ExtracurricularID,
		&item.ExtracurricularName,
		&item.StudentID,
		&item.StudentNIS,
		&item.StudentFullName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return ExtracurricularMember{}, err
	}
	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item ExtracurricularMember) error {
	if err := ensureExtracurricularExists(ctx, tx, item.ExtracurricularID); err != nil {
		return err
	}
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return err
	}
	return nil
}

func ensureExtracurricularExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM extracurriculars WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check extracurricular existence: %w", err)
	}
	if !exists {
		return ErrExtracurricularNotFound
	}
	return nil
}

func ensureStudentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM students WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student existence: %w", err)
	}
	if !exists {
		return ErrStudentNotFound
	}
	return nil
}

func ensureAcademicYearExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM academic_years WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check academic year existence: %w", err)
	}
	if !exists {
		return ErrAcademicYearNotFound
	}
	return nil
}

func ensureMemberExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM extracurricular_members WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check extracurricular member existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}
	if strings.Contains(strings.ToLower(mysqlErr.Message), "uk_extracurricular_members_active_scope") {
		return ErrDuplicateScope
	}
	return fmt.Errorf("duplicate extracurricular membership data: %w", err)
}
