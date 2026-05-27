package internshiplogs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound           = errors.New("internship log not found")
	ErrInternshipNotFound = errors.New("internship not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, internshipID uint64) ([]InternshipLog, error) {
	query := `
		SELECT
			il.id,
			il.internship_id,
			st.full_name,
			co.name,
			il.log_date,
			il.activity,
			il.notes,
			il.supervisor_name,
			il.created_at,
			il.updated_at,
			il.deleted_at
		FROM internship_logs il
		INNER JOIN internships ip ON ip.id = il.internship_id
		INNER JOIN students st ON st.id = ip.student_id
		INNER JOIN companies co ON co.id = ip.company_id
		WHERE il.deleted_at IS NULL
		  AND ip.deleted_at IS NULL
		  AND st.deleted_at IS NULL
		  AND co.deleted_at IS NULL
	`

	args := make([]any, 0, 4)
	if search != "" {
		query += " AND (st.full_name LIKE ? OR co.name LIKE ? OR il.activity LIKE ? OR il.supervisor_name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	if internshipID > 0 {
		query += " AND il.internship_id = ?"
		args = append(args, internshipID)
	}

	query += " ORDER BY il.log_date DESC, il.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query internship logs: %w", err)
	}
	defer rows.Close()

	items := make([]InternshipLog, 0)
	for rows.Next() {
		var item InternshipLog
		var notes, supervisorName sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.InternshipID,
			&item.StudentName,
			&item.CompanyName,
			&item.LogDate,
			&item.Activity,
			&notes,
			&supervisorName,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan internship log: %w", err)
		}
		if notes.Valid {
			item.Notes = notes.String
		}
		if supervisorName.Valid {
			item.SupervisorName = supervisorName.String
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate internship logs: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*InternshipLog, error) {
	const query = `
		SELECT
			il.id,
			il.internship_id,
			st.full_name,
			co.name,
			il.log_date,
			il.activity,
			il.notes,
			il.supervisor_name,
			il.created_at,
			il.updated_at,
			il.deleted_at
		FROM internship_logs il
		INNER JOIN internships ip ON ip.id = il.internship_id
		INNER JOIN students st ON st.id = ip.student_id
		INNER JOIN companies co ON co.id = ip.company_id
		WHERE il.id = ?
		  AND il.deleted_at IS NULL
		LIMIT 1
	`

	var item InternshipLog
	var notes, supervisorName sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.InternshipID,
		&item.StudentName,
		&item.CompanyName,
		&item.LogDate,
		&item.Activity,
		&notes,
		&supervisorName,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get internship log by id: %w", err)
	}
	if notes.Valid {
		item.Notes = notes.String
	}
	if supervisorName.Valid {
		item.SupervisorName = supervisorName.String
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item InternshipLog) (*InternshipLog, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create internship log transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureInternshipExists(ctx, tx, item.InternshipID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO internship_logs (internship_id, log_date, activity, notes, supervisor_name)
		VALUES (?, ?, ?, ?, ?)
	`, item.InternshipID, item.LogDate, item.Activity, nullableString(item.Notes), nullableString(item.SupervisorName))
	if err != nil {
		return nil, fmt.Errorf("insert internship log: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted internship log id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create internship log transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item InternshipLog) (*InternshipLog, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update internship log transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureInternshipExists(ctx, tx, item.InternshipID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE internship_logs
		SET internship_id = ?, log_date = ?, activity = ?, notes = ?, supervisor_name = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.InternshipID, item.LogDate, item.Activity, nullableString(item.Notes), nullableString(item.SupervisorName), id)
	if err != nil {
		return nil, fmt.Errorf("update internship log: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated internship log affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update internship log transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE internship_logs
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete internship log: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted internship log affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func ensureInternshipExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM internships WHERE id = ? AND deleted_at IS NULL)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check internship existence: %w", err)
	}
	if !exists {
		return ErrInternshipNotFound
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
