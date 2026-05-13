package teachers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound       = errors.New("teacher not found")
	ErrDuplicateNIP   = errors.New("teacher nip already exists")
	ErrDuplicateNUPTK = errors.New("teacher nuptk already exists")
	ErrDuplicateEmail = errors.New("teacher email already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, gender, status string) ([]Teacher, error) {
	query := `
		SELECT
			id,
			nip,
			nuptk,
			full_name,
			gender,
			address,
			phone,
			email,
			employment_status,
			position,
			photo_url,
			status,
			created_at,
			updated_at,
			deleted_at
		FROM teachers
		WHERE deleted_at IS NULL
	`

	args := make([]any, 0, 6)
	if search != "" {
		query += " AND (full_name LIKE ? OR nip LIKE ? OR nuptk LIKE ? OR email LIKE ? OR phone LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern)
	}
	if gender != "" {
		query += " AND gender = ?"
		args = append(args, gender)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	query += " ORDER BY full_name ASC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query teachers: %w", err)
	}
	defer rows.Close()

	items := make([]Teacher, 0)
	for rows.Next() {
		item, err := scanTeacher(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate teachers: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Teacher, error) {
	const query = `
		SELECT
			id,
			nip,
			nuptk,
			full_name,
			gender,
			address,
			phone,
			email,
			employment_status,
			position,
			photo_url,
			status,
			created_at,
			updated_at,
			deleted_at
		FROM teachers
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanTeacher(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get teacher by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Teacher) (*Teacher, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO teachers (
			nip,
			nuptk,
			full_name,
			gender,
			address,
			phone,
			email,
			employment_status,
			position,
			photo_url,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, nullableString(item.NIP), nullableString(item.NUPTK), item.FullName, nullableString(item.Gender), nullableString(item.Address), nullableString(item.Phone), nullableString(item.Email), nullableString(item.EmploymentStatus), nullableString(item.Position), nullableString(item.PhotoURL), item.Status)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert teacher: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted teacher id: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Teacher) (*Teacher, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE teachers
		SET nip = ?, nuptk = ?, full_name = ?, gender = ?, address = ?, phone = ?, email = ?, employment_status = ?, position = ?, photo_url = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, nullableString(item.NIP), nullableString(item.NUPTK), item.FullName, nullableString(item.Gender), nullableString(item.Address), nullableString(item.Phone), nullableString(item.Email), nullableString(item.EmploymentStatus), nullableString(item.Position), nullableString(item.PhotoURL), item.Status, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update teacher: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated teacher affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE teachers
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete teacher: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted teacher affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTeacher(s scanner) (Teacher, error) {
	var item Teacher
	var nip sql.NullString
	var nuptk sql.NullString
	var gender sql.NullString
	var address sql.NullString
	var phone sql.NullString
	var email sql.NullString
	var employmentStatus sql.NullString
	var position sql.NullString
	var photoURL sql.NullString
	err := s.Scan(
		&item.ID,
		&nip,
		&nuptk,
		&item.FullName,
		&gender,
		&address,
		&phone,
		&email,
		&employmentStatus,
		&position,
		&photoURL,
		&item.Status,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return Teacher{}, err
	}

	if nip.Valid {
		item.NIP = nip.String
	}
	if nuptk.Valid {
		item.NUPTK = nuptk.String
	}
	if gender.Valid {
		item.Gender = gender.String
	}
	if address.Valid {
		item.Address = address.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if email.Valid {
		item.Email = email.String
	}
	if employmentStatus.Valid {
		item.EmploymentStatus = employmentStatus.String
	}
	if position.Valid {
		item.Position = position.String
	}
	if photoURL.Valid {
		item.PhotoURL = photoURL.String
	}

	return item, nil
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
	case strings.Contains(message, "uk_teachers_nip"), strings.Contains(message, "nip"):
		return ErrDuplicateNIP
	case strings.Contains(message, "uk_teachers_nuptk"), strings.Contains(message, "nuptk"):
		return ErrDuplicateNUPTK
	case strings.Contains(message, "uk_teachers_email"), strings.Contains(message, "email"):
		return ErrDuplicateEmail
	default:
		return fmt.Errorf("duplicate teacher data: %w", err)
	}
}
