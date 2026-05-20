package internships

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound             = errors.New("internship not found")
	ErrStudentNotFound      = errors.New("student not found")
	ErrCompanyNotFound      = errors.New("company not found")
	ErrAcademicYearNotFound = errors.New("academic year not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, status string, studentID, companyID, academicYearID uint64) ([]Internship, error) {
	query := `
		SELECT
			i.id,
			i.student_id,
			s.nis,
			s.full_name,
			i.company_id,
			c.name,
			i.academic_year_id,
			ay.name,
			i.start_date,
			i.end_date,
			i.mentor_name,
			i.status,
			i.created_at,
			i.updated_at,
			i.deleted_at
		FROM internships i
		INNER JOIN students s ON s.id = i.student_id
		INNER JOIN companies c ON c.id = i.company_id
		INNER JOIN academic_years ay ON ay.id = i.academic_year_id
		WHERE i.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
	`
	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR c.name LIKE ? OR i.mentor_name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if status != "" {
		query += " AND i.status = ?"
		args = append(args, status)
	}
	if studentID > 0 {
		query += " AND i.student_id = ?"
		args = append(args, studentID)
	}
	if companyID > 0 {
		query += " AND i.company_id = ?"
		args = append(args, companyID)
	}
	if academicYearID > 0 {
		query += " AND i.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	query += " ORDER BY ay.start_date DESC, c.name ASC, s.full_name ASC, i.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query internships: %w", err)
	}
	defer rows.Close()

	items := make([]Internship, 0)
	for rows.Next() {
		item, err := scanInternship(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate internships: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Internship, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			i.id,
			i.student_id,
			s.nis,
			s.full_name,
			i.company_id,
			c.name,
			i.academic_year_id,
			ay.name,
			i.start_date,
			i.end_date,
			i.mentor_name,
			i.status,
			i.created_at,
			i.updated_at,
			i.deleted_at
		FROM internships i
		INNER JOIN students s ON s.id = i.student_id
		INNER JOIN companies c ON c.id = i.company_id
		INNER JOIN academic_years ay ON ay.id = i.academic_year_id
		WHERE i.id = ?
		  AND i.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		LIMIT 1
	`, id)
	item, err := scanInternship(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get internship by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Internship) (*Internship, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create internship transaction: %w", err)
	}
	defer tx.Rollback()
	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO internships (student_id, company_id, academic_year_id, start_date, end_date, mentor_name, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.CompanyID, item.AcademicYearID, nullableDate(item.StartDate), nullableDate(item.EndDate), nullableString(item.MentorName), item.Status)
	if err != nil {
		return nil, fmt.Errorf("insert internship: %w", err)
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted internship id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create internship transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Internship) (*Internship, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update internship transaction: %w", err)
	}
	defer tx.Rollback()
	if err := ensureInternshipExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE internships
		SET student_id = ?, company_id = ?, academic_year_id = ?, start_date = ?, end_date = ?, mentor_name = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.CompanyID, item.AcademicYearID, nullableDate(item.StartDate), nullableDate(item.EndDate), nullableString(item.MentorName), item.Status, id)
	if err != nil {
		return nil, fmt.Errorf("update internship: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated internship affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update internship transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE internships
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete internship: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted internship affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanInternship(s scanner) (Internship, error) {
	var item Internship
	var startDate sql.NullTime
	var endDate sql.NullTime
	var mentorName sql.NullString
	err := s.Scan(&item.ID, &item.StudentID, &item.StudentNIS, &item.StudentFullName, &item.CompanyID, &item.CompanyName, &item.AcademicYearID, &item.AcademicYearName, &startDate, &endDate, &mentorName, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		return Internship{}, err
	}
	if startDate.Valid {
		value := startDate.Time
		item.StartDate = &value
	}
	if endDate.Valid {
		value := endDate.Time
		item.EndDate = &value
	}
	if mentorName.Valid {
		item.MentorName = mentorName.String
	}
	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item Internship) error {
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return err
	}
	if err := ensureCompanyExists(ctx, tx, item.CompanyID); err != nil {
		return err
	}
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return err
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

func ensureCompanyExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM companies WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check company existence: %w", err)
	}
	if !exists {
		return ErrCompanyNotFound
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

func ensureInternshipExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM internships WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check internship existence: %w", err)
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
