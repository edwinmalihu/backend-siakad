package assessmentcomponents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound            = errors.New("assessment component not found")
	ErrSubjectNotFound     = errors.New("subject not found")
	ErrAcademicYearNotFound = errors.New("academic year not found")
	ErrSemesterNotFound    = errors.New("semester not found")
	ErrSemesterMismatch    = errors.New("semester does not belong to the selected academic year")
	ErrDuplicateName       = errors.New("assessment component name already exists for this scope")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, subjectID, academicYearID, semesterID uint64) ([]AssessmentComponent, error) {
	query := `
		SELECT
			ac.id,
			ac.subject_id,
			s.code,
			s.name,
			ac.academic_year_id,
			ay.name,
			ac.semester_id,
			sem.code,
			sem.name,
			ac.name,
			ac.weight,
			ac.created_at,
			ac.updated_at,
			ac.deleted_at
		FROM assessment_components ac
		INNER JOIN subjects s ON s.id = ac.subject_id
		INNER JOIN academic_years ay ON ay.id = ac.academic_year_id
		INNER JOIN semesters sem ON sem.id = ac.semester_id
		WHERE ac.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (ac.name LIKE ? OR s.code LIKE ? OR s.name LIKE ? OR ay.name LIKE ? OR sem.code LIKE ? OR sem.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if subjectID > 0 {
		query += " AND ac.subject_id = ?"
		args = append(args, subjectID)
	}

	if academicYearID > 0 {
		query += " AND ac.academic_year_id = ?"
		args = append(args, academicYearID)
	}

	if semesterID > 0 {
		query += " AND ac.semester_id = ?"
		args = append(args, semesterID)
	}

	query += " ORDER BY ay.name DESC, sem.code ASC, s.code ASC, ac.name ASC, ac.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query assessment components: %w", err)
	}
	defer rows.Close()

	items := make([]AssessmentComponent, 0)
	for rows.Next() {
		var item AssessmentComponent
		if err := rows.Scan(
			&item.ID,
			&item.SubjectID,
			&item.SubjectCode,
			&item.SubjectName,
			&item.AcademicYearID,
			&item.AcademicYearName,
			&item.SemesterID,
			&item.SemesterCode,
			&item.SemesterName,
			&item.Name,
			&item.Weight,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan assessment component: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assessment components: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*AssessmentComponent, error) {
	const query = `
		SELECT
			ac.id,
			ac.subject_id,
			s.code,
			s.name,
			ac.academic_year_id,
			ay.name,
			ac.semester_id,
			sem.code,
			sem.name,
			ac.name,
			ac.weight,
			ac.created_at,
			ac.updated_at,
			ac.deleted_at
		FROM assessment_components ac
		INNER JOIN subjects s ON s.id = ac.subject_id
		INNER JOIN academic_years ay ON ay.id = ac.academic_year_id
		INNER JOIN semesters sem ON sem.id = ac.semester_id
		WHERE ac.id = ?
		  AND ac.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		LIMIT 1
	`

	var item AssessmentComponent
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.SubjectID,
		&item.SubjectCode,
		&item.SubjectName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&item.Name,
		&item.Weight,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get assessment component by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item AssessmentComponent) (*AssessmentComponent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create assessment component transaction: %w", err)
	}
	defer tx.Rollback()

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
		INSERT INTO assessment_components (subject_id, academic_year_id, semester_id, name, weight)
		VALUES (?, ?, ?, ?, ?)
	`, item.SubjectID, item.AcademicYearID, item.SemesterID, item.Name, item.Weight)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert assessment component: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted assessment component id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create assessment component transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item AssessmentComponent) (*AssessmentComponent, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update assessment component transaction: %w", err)
	}
	defer tx.Rollback()

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
		UPDATE assessment_components
		SET subject_id = ?, academic_year_id = ?, semester_id = ?, name = ?, weight = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.SubjectID, item.AcademicYearID, item.SemesterID, item.Name, item.Weight, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update assessment component: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated assessment component affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update assessment component transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE assessment_components
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete assessment component: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted assessment component affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
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

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}

	message := strings.ToLower(mysqlErr.Message)
	switch {
	case strings.Contains(message, "uk_assessment_components_scope"):
		return ErrDuplicateName
	default:
		return fmt.Errorf("duplicate assessment component data: %w", err)
	}
}
