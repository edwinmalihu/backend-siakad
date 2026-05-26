package studentassessments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound                    = errors.New("student assessment not found")
	ErrStudentNotFound             = errors.New("student not found")
	ErrClassNotFound               = errors.New("class not found")
	ErrSubjectNotFound             = errors.New("subject not found")
	ErrAssessmentComponentNotFound = errors.New("assessment component not found")
	ErrTeacherNotFound             = errors.New("teacher not found")
	ErrAcademicYearNotFound        = errors.New("academic year not found")
	ErrSemesterNotFound            = errors.New("semester not found")
	ErrSemesterMismatch            = errors.New("semester does not belong to the selected academic year")
	ErrComponentMismatch           = errors.New("assessment component does not belong to the selected subject and semester")
	ErrDuplicateScope              = errors.New("student assessment already exists for this scope")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, studentID, classID, subjectID, assessmentComponentID, academicYearID, semesterID uint64) ([]StudentAssessment, error) {
	query := `
		SELECT
			sa.id,
			sa.student_id,
			st.full_name,
			st.nis,
			sa.class_id,
			c.name,
			sa.subject_id,
			sub.code,
			sub.name,
			sa.assessment_component_id,
			ac.name,
			ac.weight,
			sa.teacher_id,
			t.full_name,
			sa.score,
			sa.academic_year_id,
			ay.name,
			sa.semester_id,
			sem.code,
			sem.name,
			sa.created_at,
			sa.updated_at,
			sa.deleted_at
		FROM student_assessments sa
		INNER JOIN students st ON st.id = sa.student_id
		INNER JOIN classes c ON c.id = sa.class_id
		INNER JOIN subjects sub ON sub.id = sa.subject_id
		INNER JOIN assessment_components ac ON ac.id = sa.assessment_component_id
		INNER JOIN teachers t ON t.id = sa.teacher_id
		INNER JOIN academic_years ay ON ay.id = sa.academic_year_id
		INNER JOIN semesters sem ON sem.id = sa.semester_id
		WHERE sa.deleted_at IS NULL
		  AND st.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND sub.deleted_at IS NULL
		  AND ac.deleted_at IS NULL
		  AND t.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
	`

	args := make([]any, 0, 12)
	if search != "" {
		query += " AND (st.full_name LIKE ? OR st.nis LIKE ? OR c.name LIKE ? OR sub.code LIKE ? OR sub.name LIKE ? OR ac.name LIKE ? OR t.full_name LIKE ? OR ay.name LIKE ? OR sem.code LIKE ? OR sem.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if studentID > 0 {
		query += " AND sa.student_id = ?"
		args = append(args, studentID)
	}
	if classID > 0 {
		query += " AND sa.class_id = ?"
		args = append(args, classID)
	}
	if subjectID > 0 {
		query += " AND sa.subject_id = ?"
		args = append(args, subjectID)
	}
	if assessmentComponentID > 0 {
		query += " AND sa.assessment_component_id = ?"
		args = append(args, assessmentComponentID)
	}
	if academicYearID > 0 {
		query += " AND sa.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	if semesterID > 0 {
		query += " AND sa.semester_id = ?"
		args = append(args, semesterID)
	}

	query += " ORDER BY ay.name DESC, sem.code ASC, c.name ASC, st.full_name ASC, sub.code ASC, ac.name ASC, sa.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query student assessments: %w", err)
	}
	defer rows.Close()

	items := make([]StudentAssessment, 0)
	for rows.Next() {
		var item StudentAssessment
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
			&item.AssessmentComponentID,
			&item.ComponentName,
			&item.ComponentWeight,
			&item.TeacherID,
			&item.TeacherName,
			&item.Score,
			&item.AcademicYearID,
			&item.AcademicYearName,
			&item.SemesterID,
			&item.SemesterCode,
			&item.SemesterName,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan student assessment: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student assessments: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*StudentAssessment, error) {
	const query = `
		SELECT
			sa.id,
			sa.student_id,
			st.full_name,
			st.nis,
			sa.class_id,
			c.name,
			sa.subject_id,
			sub.code,
			sub.name,
			sa.assessment_component_id,
			ac.name,
			ac.weight,
			sa.teacher_id,
			t.full_name,
			sa.score,
			sa.academic_year_id,
			ay.name,
			sa.semester_id,
			sem.code,
			sem.name,
			sa.created_at,
			sa.updated_at,
			sa.deleted_at
		FROM student_assessments sa
		INNER JOIN students st ON st.id = sa.student_id
		INNER JOIN classes c ON c.id = sa.class_id
		INNER JOIN subjects sub ON sub.id = sa.subject_id
		INNER JOIN assessment_components ac ON ac.id = sa.assessment_component_id
		INNER JOIN teachers t ON t.id = sa.teacher_id
		INNER JOIN academic_years ay ON ay.id = sa.academic_year_id
		INNER JOIN semesters sem ON sem.id = sa.semester_id
		WHERE sa.id = ?
		  AND sa.deleted_at IS NULL
		  AND st.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND sub.deleted_at IS NULL
		  AND ac.deleted_at IS NULL
		  AND t.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		LIMIT 1
	`

	var item StudentAssessment
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
		&item.AssessmentComponentID,
		&item.ComponentName,
		&item.ComponentWeight,
		&item.TeacherID,
		&item.TeacherName,
		&item.Score,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get student assessment by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item StudentAssessment) (*StudentAssessment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create student assessment transaction: %w", err)
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
	if err := ensureAssessmentComponentExists(ctx, tx, item.AssessmentComponentID); err != nil {
		return nil, err
	}
	if err := ensureTeacherExists(ctx, tx, item.TeacherID); err != nil {
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
	if err := ensureComponentMatchesSubjectAndSemester(ctx, tx, item.AssessmentComponentID, item.SubjectID, item.SemesterID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO student_assessments (student_id, class_id, subject_id, assessment_component_id, teacher_id, score, academic_year_id, semester_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.ClassID, item.SubjectID, item.AssessmentComponentID, item.TeacherID, item.Score, item.AcademicYearID, item.SemesterID)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert student assessment: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted student assessment id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create student assessment transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item StudentAssessment) (*StudentAssessment, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update student assessment transaction: %w", err)
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
	if err := ensureAssessmentComponentExists(ctx, tx, item.AssessmentComponentID); err != nil {
		return nil, err
	}
	if err := ensureTeacherExists(ctx, tx, item.TeacherID); err != nil {
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
	if err := ensureComponentMatchesSubjectAndSemester(ctx, tx, item.AssessmentComponentID, item.SubjectID, item.SemesterID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE student_assessments
		SET student_id = ?, class_id = ?, subject_id = ?, assessment_component_id = ?, teacher_id = ?, score = ?, academic_year_id = ?, semester_id = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.ClassID, item.SubjectID, item.AssessmentComponentID, item.TeacherID, item.Score, item.AcademicYearID, item.SemesterID, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update student assessment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated student assessment affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update student assessment transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE student_assessments
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete student assessment: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted student assessment affected rows: %w", err)
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

func ensureAssessmentComponentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM assessment_components WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check assessment component existence: %w", err)
	}
	if !exists {
		return ErrAssessmentComponentNotFound
	}
	return nil
}

func ensureTeacherExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM teachers WHERE id = ? AND deleted_at IS NULL)
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

func ensureComponentMatchesSubjectAndSemester(ctx context.Context, tx *sql.Tx, componentID, subjectID, semesterID uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM assessment_components
			WHERE id = ? AND subject_id = ? AND semester_id = ? AND deleted_at IS NULL
		)
	`, componentID, subjectID, semesterID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check component-subject-semester match: %w", err)
	}
	if !exists {
		return ErrComponentMismatch
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
	case strings.Contains(message, "uk_student_assessments_scope"):
		return ErrDuplicateScope
	default:
		return fmt.Errorf("duplicate student assessment data: %w", err)
	}
}
