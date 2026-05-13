package homeroomassignments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound                     = errors.New("homeroom assignment not found")
	ErrTeacherNotFound              = errors.New("teacher not found")
	ErrClassNotFound                = errors.New("class not found")
	ErrAcademicYearNotFound         = errors.New("academic year not found")
	ErrSemesterNotFound             = errors.New("semester not found")
	ErrSemesterAcademicYearMismatch = errors.New("semester does not belong to the selected academic year")
	ErrClassAcademicYearMismatch    = errors.New("class does not belong to the selected academic year")
	ErrDuplicateClassScope          = errors.New("homeroom assignment already exists for the selected class, academic year, and semester")
	ErrTeacherAlreadyAssigned       = errors.New("teacher is already assigned as homeroom teacher for the selected academic year and semester")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, teacherID, classID, academicYearID, semesterID uint64) ([]HomeroomAssignment, error) {
	query := `
		SELECT
			ha.id,
			ha.teacher_id,
			COALESCE(t.nip, ''),
			t.full_name,
			ha.class_id,
			c.name,
			ha.academic_year_id,
			ay.name,
			ha.semester_id,
			sem.code,
			sem.name,
			ha.created_at,
			ha.updated_at,
			ha.deleted_at
		FROM homeroom_assignments ha
		INNER JOIN teachers t ON t.id = ha.teacher_id
		INNER JOIN classes c ON c.id = ha.class_id
		INNER JOIN academic_years ay ON ay.id = ha.academic_year_id
		INNER JOIN semesters sem ON sem.id = ha.semester_id
		WHERE ha.deleted_at IS NULL
		  AND t.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
	`

	args := make([]any, 0, 9)
	if search != "" {
		query += " AND (t.full_name LIKE ? OR t.nip LIKE ? OR c.name LIKE ? OR ay.name LIKE ? OR sem.code LIKE ? OR sem.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	if teacherID > 0 {
		query += " AND ha.teacher_id = ?"
		args = append(args, teacherID)
	}
	if classID > 0 {
		query += " AND ha.class_id = ?"
		args = append(args, classID)
	}
	if academicYearID > 0 {
		query += " AND ha.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	if semesterID > 0 {
		query += " AND ha.semester_id = ?"
		args = append(args, semesterID)
	}

	query += " ORDER BY ay.start_date DESC, sem.id DESC, c.name ASC, t.full_name ASC, ha.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query homeroom assignments: %w", err)
	}
	defer rows.Close()

	items := make([]HomeroomAssignment, 0)
	for rows.Next() {
		item, err := scanHomeroomAssignment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate homeroom assignments: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*HomeroomAssignment, error) {
	const query = `
		SELECT
			ha.id,
			ha.teacher_id,
			COALESCE(t.nip, ''),
			t.full_name,
			ha.class_id,
			c.name,
			ha.academic_year_id,
			ay.name,
			ha.semester_id,
			sem.code,
			sem.name,
			ha.created_at,
			ha.updated_at,
			ha.deleted_at
		FROM homeroom_assignments ha
		INNER JOIN teachers t ON t.id = ha.teacher_id
		INNER JOIN classes c ON c.id = ha.class_id
		INNER JOIN academic_years ay ON ay.id = ha.academic_year_id
		INNER JOIN semesters sem ON sem.id = ha.semester_id
		WHERE ha.id = ?
		  AND ha.deleted_at IS NULL
		  AND t.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanHomeroomAssignment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get homeroom assignment by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item HomeroomAssignment) (*HomeroomAssignment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create homeroom assignment transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item, 0); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO homeroom_assignments (teacher_id, class_id, academic_year_id, semester_id)
		VALUES (?, ?, ?, ?)
	`, item.TeacherID, item.ClassID, item.AcademicYearID, item.SemesterID)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert homeroom assignment: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted homeroom assignment id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create homeroom assignment transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item HomeroomAssignment) (*HomeroomAssignment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update homeroom assignment transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAssignmentExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item, id); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE homeroom_assignments
		SET teacher_id = ?, class_id = ?, academic_year_id = ?, semester_id = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.TeacherID, item.ClassID, item.AcademicYearID, item.SemesterID, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update homeroom assignment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated homeroom assignment affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update homeroom assignment transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE homeroom_assignments
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete homeroom assignment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted homeroom assignment affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanHomeroomAssignment(s scanner) (HomeroomAssignment, error) {
	var item HomeroomAssignment
	err := s.Scan(
		&item.ID,
		&item.TeacherID,
		&item.TeacherNIP,
		&item.TeacherFullName,
		&item.ClassID,
		&item.ClassName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return HomeroomAssignment{}, err
	}

	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item HomeroomAssignment, excludeID uint64) error {
	if err := ensureTeacherExists(ctx, tx, item.TeacherID); err != nil {
		return err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return err
	}
	if err := ensureSemesterMatchesAcademicYear(ctx, tx, item.SemesterID, item.AcademicYearID); err != nil {
		return err
	}
	if err := ensureClassMatchesAcademicYear(ctx, tx, item.ClassID, item.AcademicYearID); err != nil {
		return err
	}
	if err := ensureTeacherAvailable(ctx, tx, item.TeacherID, item.AcademicYearID, item.SemesterID, excludeID); err != nil {
		return err
	}
	return nil
}

func ensureTeacherExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM teachers WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check teacher existence: %w", err)
	}
	if !exists {
		return ErrTeacherNotFound
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
		return fmt.Errorf("check semester existence: %w", err)
	}
	if semesterYearID != academicYearID {
		return ErrSemesterAcademicYearMismatch
	}
	return nil
}

func ensureClassMatchesAcademicYear(ctx context.Context, tx *sql.Tx, classID, academicYearID uint64) error {
	var classYearID uint64
	err := tx.QueryRowContext(ctx, `
		SELECT academic_year_id
		FROM classes
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, classID).Scan(&classYearID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrClassNotFound
	}
	if err != nil {
		return fmt.Errorf("check class existence: %w", err)
	}
	if classYearID != academicYearID {
		return ErrClassAcademicYearMismatch
	}
	return nil
}

func ensureTeacherAvailable(ctx context.Context, tx *sql.Tx, teacherID, academicYearID, semesterID, excludeID uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM homeroom_assignments
			WHERE teacher_id = ?
			  AND academic_year_id = ?
			  AND semester_id = ?
			  AND deleted_at IS NULL
			  AND (? = 0 OR id <> ?)
		)
	`, teacherID, academicYearID, semesterID, excludeID, excludeID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check teacher homeroom availability: %w", err)
	}
	if exists {
		return ErrTeacherAlreadyAssigned
	}
	return nil
}

func ensureAssignmentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM homeroom_assignments WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check homeroom assignment existence: %w", err)
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

	message := strings.ToLower(mysqlErr.Message)
	switch {
	case strings.Contains(message, "uk_homeroom_assignments_active_scope"), strings.Contains(message, "uk_homeroom_assignments_scope"), strings.Contains(message, "active_class_id"):
		return ErrDuplicateClassScope
	default:
		return fmt.Errorf("duplicate homeroom assignment data: %w", err)
	}
}
