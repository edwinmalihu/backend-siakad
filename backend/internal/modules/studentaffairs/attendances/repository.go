package attendances

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
	ErrNotFound            = errors.New("attendance not found")
	ErrStudentNotFound     = errors.New("student not found")
	ErrClassNotFound       = errors.New("class not found")
	ErrRecordedByNotFound  = errors.New("recorded by user not found")
	ErrDuplicateScope      = errors.New("attendance already exists for the selected student, class, and date")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, status, attendanceDate string, studentID, classID uint64) ([]Attendance, error) {
	query := `
		SELECT
			a.id,
			a.student_id,
			s.nis,
			s.full_name,
			a.class_id,
			c.name,
			d.code,
			d.name,
			gl.code,
			gl.name,
			a.attendance_date,
			a.status,
			a.notes,
			a.recorded_by,
			COALESCE(up.full_name, ''),
			a.created_at,
			a.updated_at,
			a.deleted_at
		FROM attendances a
		INNER JOIN students s ON s.id = a.student_id
		INNER JOIN classes c ON c.id = a.class_id
		INNER JOIN departments d ON d.id = c.department_id
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id
		LEFT JOIN user_profiles up ON up.user_id = a.recorded_by AND up.deleted_at IS NULL
		WHERE a.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
	`

	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.full_name LIKE ? OR c.name LIKE ? OR d.code LIKE ? OR d.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern)
	}
	if status != "" {
		query += " AND a.status = ?"
		args = append(args, status)
	}
	if attendanceDate != "" {
		query += " AND a.attendance_date = ?"
		args = append(args, attendanceDate)
	}
	if studentID > 0 {
		query += " AND a.student_id = ?"
		args = append(args, studentID)
	}
	if classID > 0 {
		query += " AND a.class_id = ?"
		args = append(args, classID)
	}

	query += " ORDER BY a.attendance_date DESC, c.name ASC, s.full_name ASC, a.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query attendances: %w", err)
	}
	defer rows.Close()

	items := make([]Attendance, 0)
	for rows.Next() {
		item, err := scanAttendance(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attendances: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Attendance, error) {
	const query = `
		SELECT
			a.id,
			a.student_id,
			s.nis,
			s.full_name,
			a.class_id,
			c.name,
			d.code,
			d.name,
			gl.code,
			gl.name,
			a.attendance_date,
			a.status,
			a.notes,
			a.recorded_by,
			COALESCE(up.full_name, ''),
			a.created_at,
			a.updated_at,
			a.deleted_at
		FROM attendances a
		INNER JOIN students s ON s.id = a.student_id
		INNER JOIN classes c ON c.id = a.class_id
		INNER JOIN departments d ON d.id = c.department_id
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id
		LEFT JOIN user_profiles up ON up.user_id = a.recorded_by AND up.deleted_at IS NULL
		WHERE a.id = ?
		  AND a.deleted_at IS NULL
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND d.deleted_at IS NULL
		  AND gl.deleted_at IS NULL
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanAttendance(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get attendance by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Attendance) (*Attendance, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create attendance transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item, 0); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO attendances (student_id, class_id, attendance_date, status, notes, recorded_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, item.StudentID, item.ClassID, item.AttendanceDate.Format("2006-01-02"), item.Status, nullableString(item.Notes), nullableUint64(item.RecordedBy))
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert attendance: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted attendance id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create attendance transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Attendance) (*Attendance, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update attendance transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAttendanceExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item, id); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE attendances
		SET student_id = ?, class_id = ?, attendance_date = ?, status = ?, notes = ?, recorded_by = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.StudentID, item.ClassID, item.AttendanceDate.Format("2006-01-02"), item.Status, nullableString(item.Notes), nullableUint64(item.RecordedBy), id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update attendance: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated attendance affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update attendance transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE attendances
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete attendance: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted attendance affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAttendance(s scanner) (Attendance, error) {
	var item Attendance
	var notes sql.NullString
	var recordedBy sql.NullInt64
	err := s.Scan(
		&item.ID,
		&item.StudentID,
		&item.StudentNIS,
		&item.StudentFullName,
		&item.ClassID,
		&item.ClassName,
		&item.DepartmentCode,
		&item.DepartmentName,
		&item.GradeLevelCode,
		&item.GradeLevelName,
		&item.AttendanceDate,
		&item.Status,
		&notes,
		&recordedBy,
		&item.RecordedByName,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return Attendance{}, err
	}
	if notes.Valid {
		item.Notes = notes.String
	}
	if recordedBy.Valid {
		value := uint64(recordedBy.Int64)
		item.RecordedBy = &value
	}
	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item Attendance, excludeID uint64) error {
	if err := ensureStudentExists(ctx, tx, item.StudentID); err != nil {
		return err
	}
	if err := ensureClassExists(ctx, tx, item.ClassID); err != nil {
		return err
	}
	if err := ensureRecordedByExists(ctx, tx, item.RecordedBy); err != nil {
		return err
	}
	if err := ensureUniqueScope(ctx, tx, item.StudentID, item.ClassID, item.AttendanceDate, excludeID); err != nil {
		return err
	}
	return nil
}

func ensureStudentExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM students WHERE id = ? AND deleted_at IS NULL
		)
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
		SELECT EXISTS(
			SELECT 1 FROM classes WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check class existence: %w", err)
	}
	if !exists {
		return ErrClassNotFound
	}
	return nil
}

func ensureRecordedByExists(ctx context.Context, tx *sql.Tx, id *uint64) error {
	if id == nil || *id == 0 {
		return nil
	}
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL
		)
	`, *id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check recorded_by user existence: %w", err)
	}
	if !exists {
		return ErrRecordedByNotFound
	}
	return nil
}

func ensureUniqueScope(ctx context.Context, tx *sql.Tx, studentID, classID uint64, attendanceDate time.Time, excludeID uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM attendances
			WHERE student_id = ?
			  AND class_id = ?
			  AND attendance_date = ?
			  AND deleted_at IS NULL
			  AND (? = 0 OR id <> ?)
		)
	`, studentID, classID, attendanceDate.Format("2006-01-02"), excludeID, excludeID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check attendance uniqueness: %w", err)
	}
	if exists {
		return ErrDuplicateScope
	}
	return nil
}

func ensureAttendanceExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM attendances WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check attendance existence: %w", err)
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
	message := strings.ToLower(mysqlErr.Message)
	if strings.Contains(message, "uk_attendances_active_student_class_date") {
		return ErrDuplicateScope
	}
	return fmt.Errorf("duplicate attendance data: %w", err)
}
