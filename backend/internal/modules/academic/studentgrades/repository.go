package studentgrades

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound           = errors.New("student grade not found")
	ErrStudentNotFound    = errors.New("student not found")
	ErrClassNotFound      = errors.New("class not found")
	ErrSubjectNotFound    = errors.New("subject not found")
	ErrAcademicYearNotFound = errors.New("academic year not found")
	ErrSemesterNotFound   = errors.New("semester not found")
	ErrSemesterMismatch   = errors.New("semester does not belong to the selected academic year")
	ErrDuplicateScope     = errors.New("student grade already exists for this scope")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, studentID, classID, subjectID, academicYearID, semesterID uint64) ([]StudentGrade, error) {
	query := `
		SELECT
			sg.id,
			sg.student_id,
			st.full_name,
			st.nis,
			sg.class_id,
			c.name,
			sg.subject_id,
			sub.code,
			sub.name,
			sg.academic_year_id,
			ay.name,
			sg.semester_id,
			sem.code,
			sem.name,
			sg.final_score,
			sg.grade_letter,
			sg.predicate,
			sg.created_at,
			sg.updated_at,
			sg.deleted_at
		FROM student_grades sg
		INNER JOIN students st ON st.id = sg.student_id
		INNER JOIN classes c ON c.id = sg.class_id
		INNER JOIN subjects sub ON sub.id = sg.subject_id
		INNER JOIN academic_years ay ON ay.id = sg.academic_year_id
		INNER JOIN semesters sem ON sem.id = sg.semester_id
		WHERE sg.deleted_at IS NULL
		  AND st.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND sub.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
	`

	args := make([]any, 0, 10)
	if search != "" {
		query += " AND (st.full_name LIKE ? OR st.nis LIKE ? OR c.name LIKE ? OR sub.code LIKE ? OR sub.name LIKE ? OR ay.name LIKE ? OR sem.code LIKE ? OR sem.name LIKE ? OR sg.grade_letter LIKE ? OR sg.predicate LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if studentID > 0 {
		query += " AND sg.student_id = ?"
		args = append(args, studentID)
	}
	if classID > 0 {
		query += " AND sg.class_id = ?"
		args = append(args, classID)
	}
	if subjectID > 0 {
		query += " AND sg.subject_id = ?"
		args = append(args, subjectID)
	}
	if academicYearID > 0 {
		query += " AND sg.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	if semesterID > 0 {
		query += " AND sg.semester_id = ?"
		args = append(args, semesterID)
	}

	query += " ORDER BY ay.name DESC, sem.code ASC, c.name ASC, st.full_name ASC, sub.code ASC, sg.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query student grades: %w", err)
	}
	defer rows.Close()

	items := make([]StudentGrade, 0)
	for rows.Next() {
		var item StudentGrade
		var gradeLetter, predicate sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.StudentID,
			&item.StudentName,
			&item.StudentNIS,
			&item.ClassID,
			&item.ClassName,
			&item.SubjectID,
			&item.SubjectCode,
			&item.SubjectName,
			&item.AcademicYearID,
			&item.AcademicYearName,
			&item.SemesterID,
			&item.SemesterCode,
			&item.SemesterName,
			&item.FinalScore,
			&gradeLetter,
			&predicate,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan student grade: %w", err)
		}
		if gradeLetter.Valid {
			item.GradeLetter = gradeLetter.String
		}
		if predicate.Valid {
			item.Predicate = predicate.String
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student grades: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*StudentGrade, error) {
	const query = `
		SELECT
			sg.id,
			sg.student_id,
			st.full_name,
			st.nis,
			sg.class_id,
			c.name,
			sg.subject_id,
			sub.code,
			sub.name,
			sg.academic_year_id,
			ay.name,
			sg.semester_id,
			sem.code,
			sem.name,
			sg.final_score,
			sg.grade_letter,
			sg.predicate,
			sg.created_at,
			sg.updated_at,
			sg.deleted_at
		FROM student_grades sg
		INNER JOIN students st ON st.id = sg.student_id
		INNER JOIN classes c ON c.id = sg.class_id
		INNER JOIN subjects sub ON sub.id = sg.subject_id
		INNER JOIN academic_years ay ON ay.id = sg.academic_year_id
		INNER JOIN semesters sem ON sem.id = sg.semester_id
		WHERE sg.id = ?
		  AND sg.deleted_at IS NULL
		  AND st.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND sub.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		LIMIT 1
	`

	var item StudentGrade
	var gradeLetter, predicate sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentName,
		&item.StudentNIS,
		&item.ClassID,
		&item.ClassName,
		&item.SubjectID,
		&item.SubjectCode,
		&item.SubjectName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&item.FinalScore,
		&gradeLetter,
		&predicate,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get student grade by id: %w", err)
	}
	if gradeLetter.Valid {
		item.GradeLetter = gradeLetter.String
	}
	if predicate.Valid {
		item.Predicate = predicate.String
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item StudentGrade) (*StudentGrade, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create student grade transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return nil, err
	}
	if err := ensureClassExists(ctx, tx, item.ClassID); err != nil {
		return nil, err
	}
	if err := ensureSubjectExists(ctx, tx, item.SubjectID); err != nil {
		return nil, err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return nil, err
	}
	if err := ensureSemesterExists(ctx, tx, item.SemesterID); err != nil {
		return nil, err
	}
	if err := ensureSemesterMatchesAcademicYear(ctx, tx, item.SemesterID, item.AcademicYearID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO student_grades (student_id, class_id, subject_id, academic_year_id, semester_id, final_score, grade_letter, predicate)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.ClassID, item.SubjectID, item.AcademicYearID, item.SemesterID, item.FinalScore, nullableString(item.GradeLetter), nullableString(item.Predicate))
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert student grade: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted student grade id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create student grade transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item StudentGrade) (*StudentGrade, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update student grade transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return nil, err
	}
	if err := ensureClassExists(ctx, tx, item.ClassID); err != nil {
		return nil, err
	}
	if err := ensureSubjectExists(ctx, tx, item.SubjectID); err != nil {
		return nil, err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return nil, err
	}
	if err := ensureSemesterExists(ctx, tx, item.SemesterID); err != nil {
		return nil, err
	}
	if err := ensureSemesterMatchesAcademicYear(ctx, tx, item.SemesterID, item.AcademicYearID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE student_grades
		SET student_id = ?, class_id = ?, subject_id = ?, academic_year_id = ?, semester_id = ?, final_score = ?, grade_letter = ?, predicate = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.ClassID, item.SubjectID, item.AcademicYearID, item.SemesterID, item.FinalScore, nullableString(item.GradeLetter), nullableString(item.Predicate), id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update student grade: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated student grade affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update student grade transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE student_grades
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete student grade: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted student grade affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func ensureStudentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM students WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check student existence: %w", err)
	}
	if !exists {
		return ErrStudentNotFound
	}
	return nil
}

func ensureClassExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM classes WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check class existence: %w", err)
	}
	if !exists {
		return ErrClassNotFound
	}
	return nil
}

func ensureSubjectExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM subjects WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check subject existence: %w", err)
	}
	if !exists {
		return ErrSubjectNotFound
	}
	return nil
}

func ensureAcademicYearExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM academic_years WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check academic year existence: %w", err)
	}
	if !exists {
		return ErrAcademicYearNotFound
	}
	return nil
}

func ensureSemesterExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM semesters WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check semester existence: %w", err)
	}
	if !exists {
		return ErrSemesterNotFound
	}
	return nil
}

func ensureSemesterMatchesAcademicYear(ctx context.Context, tx *sql.Tx, semesterID, academicYearID uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM semesters WHERE id = ? AND academic_year_id = ? AND deleted_at IS NULL
		)
	`, semesterID, academicYearID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check semester-academic year match: %w", err)
	}
	if !exists {
		return ErrSemesterMismatch
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
	case strings.Contains(message, "uk_student_grades_scope"):
		return ErrDuplicateScope
	default:
		return fmt.Errorf("duplicate student grade data: %w", err)
	}
}
