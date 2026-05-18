package students

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
	ErrNotFound       = errors.New("student not found")
	ErrDuplicateNIS   = errors.New("student nis already exists")
	ErrDuplicateNISN  = errors.New("student nisn already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, gender, status string, entryYear int) ([]Student, error) {
	query := `
		SELECT
			id,
			nis,
			nisn,
			full_name,
			gender,
			birth_place,
			birth_date,
			address,
			phone,
			CAST(entry_year AS UNSIGNED),
			status,
			created_at,
			updated_at,
			deleted_at
		FROM students
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 6)
	if search != "" {
		query += " AND (nis LIKE ? OR nisn LIKE ? OR full_name LIKE ? OR phone LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if gender != "" {
		query += " AND gender = ?"
		args = append(args, gender)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if entryYear > 0 {
		query += " AND entry_year = ?"
		args = append(args, entryYear)
	}

	query += " ORDER BY full_name ASC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query students: %w", err)
	}
	defer rows.Close()

	items := make([]Student, 0)
	for rows.Next() {
		item, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate students: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Student, error) {
	const query = `
		SELECT
			id,
			nis,
			nisn,
			full_name,
			gender,
			birth_place,
			birth_date,
			address,
			phone,
			CAST(entry_year AS UNSIGNED),
			status,
			created_at,
			updated_at,
			deleted_at
		FROM students
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanStudent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get student by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Student) (*Student, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO students (
			nis,
			nisn,
			full_name,
			gender,
			birth_place,
			birth_date,
			address,
			phone,
			entry_year,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		item.NIS,
		nullableString(item.NISN),
		item.FullName,
		item.Gender,
		nullableString(item.BirthPlace),
		nullableDate(item.BirthDate),
		nullableString(item.Address),
		nullableString(item.Phone),
		item.EntryYear,
		item.Status,
	)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert student: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted student id: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Student) (*Student, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE students
		SET nis = ?, nisn = ?, full_name = ?, gender = ?, birth_place = ?, birth_date = ?, address = ?, phone = ?, entry_year = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`,
		item.NIS,
		nullableString(item.NISN),
		item.FullName,
		item.Gender,
		nullableString(item.BirthPlace),
		nullableDate(item.BirthDate),
		nullableString(item.Address),
		nullableString(item.Phone),
		item.EntryYear,
		item.Status,
		id,
	)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update student: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated student affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE students
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete student: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted student affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanStudent(s scanner) (Student, error) {
	var item Student
	var nisn sql.NullString
	var birthPlace sql.NullString
	var birthDate sql.NullTime
	var address sql.NullString
	var phone sql.NullString

	err := s.Scan(
		&item.ID,
		&item.NIS,
		&nisn,
		&item.FullName,
		&item.Gender,
		&birthPlace,
		&birthDate,
		&address,
		&phone,
		&item.EntryYear,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return Student{}, err
	}

	if nisn.Valid {
		item.NISN = nisn.String
	}
	if birthPlace.Valid {
		item.BirthPlace = birthPlace.String
	}
	if birthDate.Valid {
		value := birthDate.Time
		item.BirthDate = &value
	}
	if address.Valid {
		item.Address = address.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}

	return item, nil
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
	switch {
	case strings.Contains(message, "uk_students_nisn"):
		return ErrDuplicateNISN
	case strings.Contains(message, "uk_students_nis"):
		return ErrDuplicateNIS
	default:
		return fmt.Errorf("duplicate student data: %w", err)
	}
}
