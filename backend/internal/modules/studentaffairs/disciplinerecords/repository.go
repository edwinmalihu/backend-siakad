package disciplinerecords

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound                   = errors.New("discipline record not found")
	ErrStudentNotFound            = errors.New("student not found")
	ErrDisciplineCategoryNotFound = errors.New("discipline category not found")
	ErrRecordedByNotFound         = errors.New("recorded by user not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, incidentDate string, studentID, disciplineCategoryID uint64) ([]DisciplineRecord, error) {
	query := `
		SELECT
			dr.id,
			dr.student_id,
			s.nis,
			s.full_name,
			dr.discipline_category_id,
			dc.name,
			dc.point,
			dr.recorded_by,
			COALESCE(up.full_name, ''),
			dr.incident_date,
			dr.description,
			dr.action_taken,
			dr.created_at,
			dr.updated_at,
			dr.deleted_at
		FROM discipline_records dr
		INNER JOIN students s ON s.id = dr.student_id
		INNER JOIN discipline_categories dc ON dc.id = dr.discipline_category_id
		LEFT JOIN user_profiles up ON up.user_id = dr.recorded_by AND up.deleted_at IS NULL
		WHERE dr.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND dc.deleted_at IS NULL
	`

	args := make([]any, 0, 7)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR dc.name LIKE ? OR dr.description LIKE ? OR dr.action_taken LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern)
	}
	if incidentDate != "" {
		query += " AND dr.incident_date = ?"
		args = append(args, incidentDate)
	}
	if studentID > 0 {
		query += " AND dr.student_id = ?"
		args = append(args, studentID)
	}
	if disciplineCategoryID > 0 {
		query += " AND dr.discipline_category_id = ?"
		args = append(args, disciplineCategoryID)
	}

	query += " ORDER BY dr.incident_date DESC, dc.point DESC, s.full_name ASC, dr.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query discipline records: %w", err)
	}
	defer rows.Close()

	items := make([]DisciplineRecord, 0)
	for rows.Next() {
		item, err := scanDisciplineRecord(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate discipline records: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*DisciplineRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			dr.id,
			dr.student_id,
			s.nis,
			s.full_name,
			dr.discipline_category_id,
			dc.name,
			dc.point,
			dr.recorded_by,
			COALESCE(up.full_name, ''),
			dr.incident_date,
			dr.description,
			dr.action_taken,
			dr.created_at,
			dr.updated_at,
			dr.deleted_at
		FROM discipline_records dr
		INNER JOIN students s ON s.id = dr.student_id
		INNER JOIN discipline_categories dc ON dc.id = dr.discipline_category_id
		LEFT JOIN user_profiles up ON up.user_id = dr.recorded_by AND up.deleted_at IS NULL
		WHERE dr.id = ?
		  AND dr.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND dc.deleted_at IS NULL
		LIMIT 1
	`, id)

	item, err := scanDisciplineRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get discipline record by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item DisciplineRecord) (*DisciplineRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create discipline record transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO discipline_records (
			student_id, discipline_category_id, recorded_by, incident_date, description, action_taken
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.DisciplineCategoryID, nullableUint64(item.RecordedBy), item.IncidentDate.Format("2006-01-02"), nullableString(item.Description), nullableString(item.ActionTaken))
	if err != nil {
		return nil, fmt.Errorf("insert discipline record: %w", err)
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted discipline record id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create discipline record transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item DisciplineRecord) (*DisciplineRecord, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update discipline record transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureDisciplineRecordExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE discipline_records
		SET student_id = ?, discipline_category_id = ?, recorded_by = ?, incident_date = ?, description = ?, action_taken = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.DisciplineCategoryID, nullableUint64(item.RecordedBy), item.IncidentDate.Format("2006-01-02"), nullableString(item.Description), nullableString(item.ActionTaken), id)
	if err != nil {
		return nil, fmt.Errorf("update discipline record: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated discipline record affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update discipline record transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE discipline_records
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete discipline record: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted discipline record affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDisciplineRecord(s scanner) (DisciplineRecord, error) {
	var item DisciplineRecord
	var recordedBy sql.NullInt64
	var description sql.NullString
	var actionTaken sql.NullString
	err := s.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentNIS,
		&item.StudentFullName,
		&item.DisciplineCategoryID,
		&item.DisciplineCategoryName,
		&item.Point,
		&recordedBy,
		&item.RecordedByName,
		&item.IncidentDate,
		&description,
		&actionTaken,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return DisciplineRecord{}, err
	}
	if recordedBy.Valid {
		value := uint64(recordedBy.Int64)
		item.RecordedBy = &value
	}
	if description.Valid {
		item.Description = description.String
	}
	if actionTaken.Valid {
		item.ActionTaken = actionTaken.String
	}
	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item DisciplineRecord) error {
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return err
	}
	if err := ensureDisciplineCategoryExists(ctx, tx, item.DisciplineCategoryID); err != nil {
		return err
	}
	if err := ensureRecordedByExists(ctx, tx, item.RecordedBy); err != nil {
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

func ensureDisciplineCategoryExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM discipline_categories WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check discipline category existence: %w", err)
	}
	if !exists {
		return ErrDisciplineCategoryNotFound
	}
	return nil
}

func ensureRecordedByExists(ctx context.Context, tx *sql.Tx, id *uint64) error {
	if id == nil || *id == 0 {
		return nil
	}
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL)`, *id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check recorded_by user existence: %w", err)
	}
	if !exists {
		return ErrRecordedByNotFound
	}
	return nil
}

func ensureDisciplineRecordExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM discipline_records WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check discipline record existence: %w", err)
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
