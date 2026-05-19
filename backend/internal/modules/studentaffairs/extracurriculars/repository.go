package extracurriculars

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound             = errors.New("extracurricular not found")
	ErrCoachTeacherNotFound = errors.New("coach teacher not found")
	ErrDuplicateName        = errors.New("extracurricular name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, coachTeacherID uint64, isActive *bool) ([]Extracurricular, error) {
	query := `
		SELECT
			e.id,
			e.coach_teacher_id,
			COALESCE(t.full_name, ''),
			e.name,
			e.description,
			e.is_active,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM extracurriculars e
		LEFT JOIN teachers t ON t.id = e.coach_teacher_id AND t.deleted_at IS NULL
		WHERE e.deleted_at IS NULL
	`

	args := make([]any, 0, 5)
	if search != "" {
		query += " AND (e.name LIKE ? OR e.description LIKE ? OR t.full_name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if coachTeacherID > 0 {
		query += " AND e.coach_teacher_id = ?"
		args = append(args, coachTeacherID)
	}
	if isActive != nil {
		query += " AND e.is_active = ?"
		args = append(args, *isActive)
	}

	query += " ORDER BY e.is_active DESC, e.name ASC, e.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query extracurriculars: %w", err)
	}
	defer rows.Close()

	items := make([]Extracurricular, 0)
	for rows.Next() {
		item, err := scanExtracurricular(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate extracurriculars: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Extracurricular, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			e.id,
			e.coach_teacher_id,
			COALESCE(t.full_name, ''),
			e.name,
			e.description,
			e.is_active,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM extracurriculars e
		LEFT JOIN teachers t ON t.id = e.coach_teacher_id AND t.deleted_at IS NULL
		WHERE e.id = ? AND e.deleted_at IS NULL
		LIMIT 1
	`, id)

	item, err := scanExtracurricular(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get extracurricular by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Extracurricular) (*Extracurricular, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create extracurricular transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureCoachTeacherExists(ctx, tx, item.CoachTeacherID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO extracurriculars (coach_teacher_id, name, description, is_active)
		VALUES (?, ?, ?, ?)
	`, nullableUint64(item.CoachTeacherID), item.Name, nullableString(item.Description), item.IsActive)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert extracurricular: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted extracurricular id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create extracurricular transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Extracurricular) (*Extracurricular, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update extracurricular transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureExtracurricularExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := ensureCoachTeacherExists(ctx, tx, item.CoachTeacherID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE extracurriculars
		SET coach_teacher_id = ?, name = ?, description = ?, is_active = ?
		WHERE id = ? AND deleted_at IS NULL
	`, nullableUint64(item.CoachTeacherID), item.Name, nullableString(item.Description), item.IsActive, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update extracurricular: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated extracurricular affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update extracurricular transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE extracurriculars
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete extracurricular: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted extracurricular affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanExtracurricular(s scanner) (Extracurricular, error) {
	var item Extracurricular
	var coachTeacherID sql.NullInt64
	var description sql.NullString
	err := s.Scan(
		&item.ID,
		&coachTeacherID,
		&item.CoachTeacherName,
		&item.Name,
		&description,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return Extracurricular{}, err
	}
	if coachTeacherID.Valid {
		value := uint64(coachTeacherID.Int64)
		item.CoachTeacherID = &value
	}
	if description.Valid {
		item.Description = description.String
	}
	return item, nil
}

func ensureCoachTeacherExists(ctx context.Context, tx *sql.Tx, id *uint64) error {
	if id == nil || *id == 0 {
		return nil
	}
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM teachers WHERE id = ? AND deleted_at IS NULL)`, *id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check coach teacher existence: %w", err)
	}
	if !exists {
		return ErrCoachTeacherNotFound
	}
	return nil
}

func ensureExtracurricularExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM extracurriculars WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check extracurricular existence: %w", err)
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

func nullableUint64(value *uint64) any {
	if value == nil || *value == 0 {
		return nil
	}
	return *value
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}
	if strings.Contains(strings.ToLower(mysqlErr.Message), "uk_extracurriculars_active_name") {
		return ErrDuplicateName
	}
	return fmt.Errorf("duplicate extracurricular data: %w", err)
}
