package studentenrollments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound                     = errors.New("student enrollment not found")
	ErrStudentNotFound              = errors.New("student not found")
	ErrClassNotFound                = errors.New("class not found")
	ErrAcademicYearNotFound         = errors.New("academic year not found")
	ErrSemesterNotFound             = errors.New("semester not found")
	ErrSemesterAcademicYearMismatch = errors.New("semester does not belong to the selected academic year")
	ErrClassAcademicYearMismatch    = errors.New("class does not belong to the selected academic year")
	ErrDuplicateScope               = errors.New("student enrollment already exists for the selected academic year and semester")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, status string, studentID, classID, academicYearID, semesterID uint64) ([]StudentEnrollment, error) {
	query := `
		SELECT
			se.id,
			se.student_id,
			s.nis,
			s.full_name,
			se.class_id,
			c.name,
			d.code,
			d.name,
			gl.code,
			gl.name,
			se.academic_year_id,
			ay.name,
			se.semester_id,
			sem.code,
			sem.name,
			se.status,
			se.created_at,
			se.updated_at,
			se.deleted_at
		FROM student_enrollments se
		INNER JOIN students s ON s.id = se.student_id
		INNER JOIN classes c ON c.id = se.class_id
		INNER JOIN departments d ON d.id = c.department_id
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id
		INNER JOIN academic_years ay ON ay.id = se.academic_year_id
		INNER JOIN semesters sem ON sem.id = se.semester_id
		WHERE se.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR c.name LIKE ? OR ay.name LIKE ? OR sem.code LIKE ? OR sem.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	if status != "" {
		query += " AND se.status = ?"
		args = append(args, status)
	}
	if studentID > 0 {
		query += " AND se.student_id = ?"
		args = append(args, studentID)
	}
	if classID > 0 {
		query += " AND se.class_id = ?"
		args = append(args, classID)
	}
	if academicYearID > 0 {
		query += " AND se.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	if semesterID > 0 {
		query += " AND se.semester_id = ?"
		args = append(args, semesterID)
	}

	query += " ORDER BY ay.start_date DESC, sem.id DESC, s.full_name ASC, se.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query student enrollments: %w", err)
	}
	defer rows.Close()

	items := make([]StudentEnrollment, 0)
	for rows.Next() {
		item, err := scanStudentEnrollment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student enrollments: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*StudentEnrollment, error) {
	const query = `
		SELECT
			se.id,
			se.student_id,
			s.nis,
			s.full_name,
			se.class_id,
			c.name,
			d.code,
			d.name,
			gl.code,
			gl.name,
			se.academic_year_id,
			ay.name,
			se.semester_id,
			sem.code,
			sem.name,
			se.status,
			se.created_at,
			se.updated_at,
			se.deleted_at
		FROM student_enrollments se
		INNER JOIN students s ON s.id = se.student_id
		INNER JOIN classes c ON c.id = se.class_id
		INNER JOIN departments d ON d.id = c.department_id
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id
		INNER JOIN academic_years ay ON ay.id = se.academic_year_id
		INNER JOIN semesters sem ON sem.id = se.semester_id
		WHERE se.id = ?
		  AND se.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanStudentEnrollment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get student enrollment by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item StudentEnrollment) (*StudentEnrollment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create student enrollment transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item, 0); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO student_enrollments (student_id, class_id, academic_year_id, semester_id, status)
		VALUES (?, ?, ?, ?, ?)
	`, item.StudentID, item.ClassID, item.AcademicYearID, item.SemesterID, item.Status)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert student enrollment: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted student enrollment id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create student enrollment transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item StudentEnrollment) (*StudentEnrollment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update student enrollment transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureEnrollmentExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item, id); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE student_enrollments
		SET student_id = ?, class_id = ?, academic_year_id = ?, semester_id = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.ClassID, item.AcademicYearID, item.SemesterID, item.Status, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update student enrollment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated student enrollment affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update student enrollment transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE student_enrollments
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete student enrollment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted student enrollment affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanStudentEnrollment(s scanner) (StudentEnrollment, error) {
	var item StudentEnrollment
	err := s.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentNIS,
		&item.StudentFullName,
		&item.ClassID,
		&item.ClassName,
		&item.DepartmentCode,
		&item.DepartmentName,
		&item.GradeLevelCode,
		&item.GradeLevelName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return StudentEnrollment{}, err
	}

	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item StudentEnrollment, excludeID uint64) error {
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
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
	if err := ensureUniqueScope(ctx, tx, item.StudentID, item.AcademicYearID, item.SemesterID, excludeID); err != nil {
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
		return fmt.Errorf("check class academic year: %w", err)
	}
	if classYearID != academicYearID {
		return ErrClassAcademicYearMismatch
	}
	return nil
}

func ensureUniqueScope(ctx context.Context, tx *sql.Tx, studentID, academicYearID, semesterID, excludeID uint64) error {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM student_enrollments
			WHERE student_id = ?
			  AND academic_year_id = ?
			  AND semester_id = ?
			  AND deleted_at IS NULL
			  AND (? = 0 OR id <> ?)
		)
	`

	var exists bool
	err := tx.QueryRowContext(ctx, query, studentID, academicYearID, semesterID, excludeID, excludeID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student enrollment uniqueness: %w", err)
	}
	if exists {
		return ErrDuplicateScope
	}
	return nil
}

func ensureEnrollmentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM student_enrollments WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student enrollment existence: %w", err)
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
	if strings.Contains(message, "uk_student_enrollments_active_scope") {
		return ErrDuplicateScope
	}

	return fmt.Errorf("duplicate student enrollment data: %w", err)
}
