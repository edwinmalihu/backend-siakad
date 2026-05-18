package studentmutations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound                     = errors.New("student mutation not found")
	ErrStudentNotFound              = errors.New("student not found")
	ErrAcademicYearNotFound         = errors.New("academic year not found")
	ErrSemesterNotFound             = errors.New("semester not found")
	ErrSemesterAcademicYearMismatch = errors.New("semester does not belong to the selected academic year")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, mutationType, status string, studentID, academicYearID, semesterID uint64) ([]StudentMutation, error) {
	query := `
		SELECT
			sm.id,
			sm.student_id,
			s.nis,
			s.full_name,
			sm.academic_year_id,
			ay.name,
			sm.semester_id,
			sem.code,
			sem.name,
			sm.mutation_type,
			sm.from_school,
			sm.to_school,
			sm.reason,
			sm.effective_date,
			sm.status,
			sm.created_at,
			sm.updated_at,
			sm.deleted_at
		FROM student_mutations sm
		INNER JOIN students s ON s.id = sm.student_id
		INNER JOIN academic_years ay ON ay.id = sm.academic_year_id
		INNER JOIN semesters sem ON sem.id = sm.semester_id
		WHERE sm.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR ay.name LIKE ? OR sem.code LIKE ? OR sm.from_school LIKE ? OR sm.to_school LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	if mutationType != "" {
		query += " AND sm.mutation_type = ?"
		args = append(args, mutationType)
	}
	if status != "" {
		query += " AND sm.status = ?"
		args = append(args, status)
	}
	if studentID > 0 {
		query += " AND sm.student_id = ?"
		args = append(args, studentID)
	}
	if academicYearID > 0 {
		query += " AND sm.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	if semesterID > 0 {
		query += " AND sm.semester_id = ?"
		args = append(args, semesterID)
	}

	query += " ORDER BY COALESCE(sm.effective_date, DATE(sm.created_at)) DESC, sm.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query student mutations: %w", err)
	}
	defer rows.Close()

	items := make([]StudentMutation, 0)
	for rows.Next() {
		item, err := scanStudentMutation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student mutations: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*StudentMutation, error) {
	const query = `
		SELECT
			sm.id,
			sm.student_id,
			s.nis,
			s.full_name,
			sm.academic_year_id,
			ay.name,
			sm.semester_id,
			sem.code,
			sem.name,
			sm.mutation_type,
			sm.from_school,
			sm.to_school,
			sm.reason,
			sm.effective_date,
			sm.status,
			sm.created_at,
			sm.updated_at,
			sm.deleted_at
		FROM student_mutations sm
		INNER JOIN students s ON s.id = sm.student_id
		INNER JOIN academic_years ay ON ay.id = sm.academic_year_id
		INNER JOIN semesters sem ON sem.id = sm.semester_id
		WHERE sm.id = ?
		  AND sm.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanStudentMutation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get student mutation by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item StudentMutation) (*StudentMutation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create student mutation transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO student_mutations (
			student_id, academic_year_id, semester_id, mutation_type, from_school, to_school, reason, effective_date, status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.AcademicYearID, item.SemesterID, item.MutationType, nullableString(item.FromSchool), nullableString(item.ToSchool), nullableString(item.Reason), nullableDate(item.EffectiveDate), item.Status)
	if err != nil {
		return nil, fmt.Errorf("insert student mutation: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted student mutation id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create student mutation transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item StudentMutation) (*StudentMutation, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update student mutation transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureMutationExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE student_mutations
		SET student_id = ?, academic_year_id = ?, semester_id = ?, mutation_type = ?, from_school = ?, to_school = ?, reason = ?, effective_date = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.AcademicYearID, item.SemesterID, item.MutationType, nullableString(item.FromSchool), nullableString(item.ToSchool), nullableString(item.Reason), nullableDate(item.EffectiveDate), item.Status, id)
	if err != nil {
		return nil, fmt.Errorf("update student mutation: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated student mutation affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update student mutation transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE student_mutations
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete student mutation: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted student mutation affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanStudentMutation(s scanner) (StudentMutation, error) {
	var item StudentMutation
	var fromSchool sql.NullString
	var toSchool sql.NullString
	var reason sql.NullString
	var effectiveDate sql.NullTime
	err := s.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentNIS,
		&item.StudentFullName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&item.MutationType,
		&fromSchool,
		&toSchool,
		&reason,
		&effectiveDate,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return StudentMutation{}, err
	}
	if fromSchool.Valid {
		item.FromSchool = fromSchool.String
	}
	if toSchool.Valid {
		item.ToSchool = toSchool.String
	}
	if reason.Valid {
		item.Reason = reason.String
	}
	if effectiveDate.Valid {
		value := effectiveDate.Time
		item.EffectiveDate = &value
	}
	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item StudentMutation) error {
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return err
	}
	if err := ensureSemesterMatchesAcademicYear(ctx, tx, item.SemesterID, item.AcademicYearID); err != nil {
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

func ensureSemesterMatchesAcademicYear(ctx context.Context, tx *sql.Tx, semesterID, academicYearID uint64) error {
	var semesterYearID uint64
	err := tx.QueryRowContext(ctx, `
		SELECT academic_year_id
		FROM semesters
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, semesterID).Scan(&semesterYearID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSemesterNotFound
	}
	if err != nil {
		return fmt.Errorf("check semester academic year: %w", err)
	}
	if semesterYearID != academicYearID {
		return ErrSemesterAcademicYearMismatch
	}
	return nil
}

func ensureMutationExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM student_mutations WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student mutation existence: %w", err)
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
