package studentgraduations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound             = errors.New("student graduation not found")
	ErrStudentNotFound      = errors.New("student not found")
	ErrAcademicYearNotFound = errors.New("academic year not found")
	ErrDuplicateScope       = errors.New("student graduation already exists for the selected academic year")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, status string, studentID, academicYearID uint64) ([]StudentGraduation, error) {
	query := `
		SELECT
			sg.id,
			sg.student_id,
			s.nis,
			s.full_name,
			sg.academic_year_id,
			ay.name,
			sg.graduation_date,
			sg.status,
			sg.notes,
			sg.created_at,
			sg.updated_at,
			sg.deleted_at
		FROM student_graduations sg
		INNER JOIN students s ON s.id = sg.student_id
		INNER JOIN academic_years ay ON ay.id = sg.academic_year_id
		WHERE sg.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
	`

	args := make([]any, 0, 6)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR ay.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if status != "" {
		query += " AND sg.status = ?"
		args = append(args, status)
	}
	if studentID > 0 {
		query += " AND sg.student_id = ?"
		args = append(args, studentID)
	}
	if academicYearID > 0 {
		query += " AND sg.academic_year_id = ?"
		args = append(args, academicYearID)
	}

	query += " ORDER BY ay.start_date DESC, COALESCE(sg.graduation_date, DATE(sg.created_at)) DESC, sg.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query student graduations: %w", err)
	}
	defer rows.Close()

	items := make([]StudentGraduation, 0)
	for rows.Next() {
		item, err := scanStudentGraduation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student graduations: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*StudentGraduation, error) {
	const query = `
		SELECT
			sg.id,
			sg.student_id,
			s.nis,
			s.full_name,
			sg.academic_year_id,
			ay.name,
			sg.graduation_date,
			sg.status,
			sg.notes,
			sg.created_at,
			sg.updated_at,
			sg.deleted_at
		FROM student_graduations sg
		INNER JOIN students s ON s.id = sg.student_id
		INNER JOIN academic_years ay ON ay.id = sg.academic_year_id
		WHERE sg.id = ?
		  AND sg.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanStudentGraduation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get student graduation by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item StudentGraduation) (*StudentGraduation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create student graduation transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item, 0); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO student_graduations (student_id, academic_year_id, graduation_date, status, notes)
		VALUES (?, ?, ?, ?, ?)
	`, item.StudentID, item.AcademicYearID, nullableDate(item.GraduationDate), item.Status, nullableString(item.Notes))
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert student graduation: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted student graduation id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create student graduation transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item StudentGraduation) (*StudentGraduation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update student graduation transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureGraduationExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item, id); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE student_graduations
		SET student_id = ?, academic_year_id = ?, graduation_date = ?, status = ?, notes = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.AcademicYearID, nullableDate(item.GraduationDate), item.Status, nullableString(item.Notes), id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update student graduation: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated student graduation affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update student graduation transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE student_graduations
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete student graduation: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted student graduation affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanStudentGraduation(s scanner) (StudentGraduation, error) {
	var item StudentGraduation
	var graduationDate sql.NullTime
	var notes sql.NullString
	err := s.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentNIS,
		&item.StudentFullName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&graduationDate,
		&item.Status,
		&notes,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return StudentGraduation{}, err
	}
	if graduationDate.Valid {
		value := graduationDate.Time
		item.GraduationDate = &value
	}
	if notes.Valid {
		item.Notes = notes.String
	}
	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item StudentGraduation, excludeID uint64) error {
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return err
	}
	if err := ensureUniqueScope(ctx, tx, item.StudentID, item.AcademicYearID, excludeID); err != nil {
		return err
	}
	return nil
}

func ensureStudentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM students WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
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

func ensureUniqueScope(ctx context.Context, tx *sql.Tx, studentID, academicYearID, excludeID uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM student_graduations
			WHERE student_id = ?
			  AND academic_year_id = ?
			  AND deleted_at IS NULL
			  AND (? = 0 OR id <> ?)
		)
	`, studentID, academicYearID, excludeID, excludeID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student graduation uniqueness: %w", err)
	}
	if exists {
		return ErrDuplicateScope
	}
	return nil
}

func ensureGraduationExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM student_graduations WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student graduation existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02")
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}
	message := strings.ToLower(mysqlErr.Message)
	if strings.Contains(message, "uk_student_graduations_active_student_year") {
		return ErrDuplicateScope
	}
	return fmt.Errorf("duplicate student graduation data: %w", err)
}
