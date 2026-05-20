package alumni

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound         = errors.New("alumnus not found")
	ErrStudentNotFound  = errors.New("student not found")
	ErrDuplicateStudent = errors.New("student is already registered as alumni")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, currentActivity string, studentID uint64, graduationYear int) ([]Alumnus, error) {
	query := `
		SELECT
			a.id,
			a.student_id,
			s.nis,
			s.full_name,
			a.graduation_year,
			a.current_activity,
			a.company_name,
			a.college_name,
			a.phone,
			a.email,
			a.created_at,
			a.updated_at,
			a.deleted_at
		FROM alumni a
		INNER JOIN students s ON s.id = a.student_id
		WHERE a.deleted_at IS NULL
		  AND s.deleted_at IS NULL
	`
	args := make([]any, 0, 6)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR a.company_name LIKE ? OR a.college_name LIKE ? OR a.email LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern)
	}
	if currentActivity != "" {
		query += " AND a.current_activity = ?"
		args = append(args, currentActivity)
	}
	if studentID > 0 {
		query += " AND a.student_id = ?"
		args = append(args, studentID)
	}
	if graduationYear > 0 {
		query += " AND a.graduation_year = ?"
		args = append(args, graduationYear)
	}
	query += " ORDER BY a.graduation_year DESC, s.full_name ASC, a.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query alumni: %w", err)
	}
	defer rows.Close()

	items := make([]Alumnus, 0)
	for rows.Next() {
		item, err := scanAlumnus(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate alumni: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Alumnus, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			a.id,
			a.student_id,
			s.nis,
			s.full_name,
			a.graduation_year,
			a.current_activity,
			a.company_name,
			a.college_name,
			a.phone,
			a.email,
			a.created_at,
			a.updated_at,
			a.deleted_at
		FROM alumni a
		INNER JOIN students s ON s.id = a.student_id
		WHERE a.id = ?
		  AND a.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		LIMIT 1
	`, id)
	item, err := scanAlumnus(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get alumnus by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Alumnus) (*Alumnus, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create alumnus transaction: %w", err)
	}
	defer tx.Rollback()
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO alumni (student_id, graduation_year, current_activity, company_name, college_name, phone, email)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.GraduationYear, nullableString(item.CurrentActivity), nullableString(item.CompanyName), nullableString(item.CollegeName), nullableString(item.Phone), nullableString(item.Email))
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert alumnus: %w", err)
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted alumnus id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create alumnus transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Alumnus) (*Alumnus, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update alumnus transaction: %w", err)
	}
	defer tx.Rollback()
	if err := ensureAlumnusExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE alumni
		SET student_id = ?, graduation_year = ?, current_activity = ?, company_name = ?, college_name = ?, phone = ?, email = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.GraduationYear, nullableString(item.CurrentActivity), nullableString(item.CompanyName), nullableString(item.CollegeName), nullableString(item.Phone), nullableString(item.Email), id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update alumnus: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated alumnus affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update alumnus transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE alumni
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete alumnus: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted alumnus affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAlumnus(s scanner) (Alumnus, error) {
	var item Alumnus
	var currentActivity sql.NullString
	var companyName sql.NullString
	var collegeName sql.NullString
	var phone sql.NullString
	var email sql.NullString
	err := s.Scan(&item.ID, &item.StudentID, &item.StudentNIS, &item.StudentFullName, &item.GraduationYear, &currentActivity, &companyName, &collegeName, &phone, &email, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		return Alumnus{}, err
	}
	if currentActivity.Valid {
		item.CurrentActivity = currentActivity.String
	}
	if companyName.Valid {
		item.CompanyName = companyName.String
	}
	if collegeName.Valid {
		item.CollegeName = collegeName.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if email.Valid {
		item.Email = email.String
	}
	return item, nil
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

func ensureAlumnusExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM alumni WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check alumnus existence: %w", err)
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

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}
	if strings.Contains(strings.ToLower(mysqlErr.Message), "uk_alumni_active_student_id") {
		return ErrDuplicateStudent
	}
	return fmt.Errorf("duplicate alumni data: %w", err)
}
