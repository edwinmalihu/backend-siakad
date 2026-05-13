package schedules

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound                     = errors.New("schedule not found")
	ErrAcademicYearNotFound         = errors.New("academic year not found")
	ErrSemesterNotFound             = errors.New("semester not found")
	ErrClassNotFound                = errors.New("class not found")
	ErrSubjectNotFound              = errors.New("subject not found")
	ErrTeacherNotFound              = errors.New("teacher not found")
	ErrRoomNotFound                 = errors.New("room not found")
	ErrSemesterAcademicYearMismatch = errors.New("semester does not belong to the selected academic year")
	ErrClassAcademicYearMismatch    = errors.New("class does not belong to the selected academic year")
	ErrSubjectScopeMismatch         = errors.New("subject does not match the selected class scope")
	ErrClassScheduleConflict        = errors.New("class schedule conflicts with another schedule")
	ErrTeacherScheduleConflict      = errors.New("teacher schedule conflicts with another schedule")
	ErrRoomScheduleConflict         = errors.New("room schedule conflicts with another schedule")
)

type Repository struct {
	db *sql.DB
}

type classScope struct {
	AcademicYearID uint64
	DepartmentID   uint64
	GradeLevelID   uint64
}

type subjectScope struct {
	DepartmentID uint64
	GradeLevelID uint64
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search string, classID, teacherID, roomID, academicYearID, semesterID uint64, dayOfWeek *uint8) ([]Schedule, error) {
	query := `
		SELECT
			s.id,
			s.class_id,
			c.name,
			s.subject_id,
			sub.code,
			sub.name,
			s.teacher_id,
			t.full_name,
			s.room_id,
			COALESCE(rm.code, ''),
			COALESCE(rm.name, ''),
			s.academic_year_id,
			ay.name,
			s.semester_id,
			sem.code,
			sem.name,
			s.day_of_week,
			TIME_FORMAT(s.start_time, '%H:%i:%s'),
			TIME_FORMAT(s.end_time, '%H:%i:%s'),
			s.created_at,
			s.updated_at,
			s.deleted_at
		FROM schedules s
		INNER JOIN classes c ON c.id = s.class_id
		INNER JOIN subjects sub ON sub.id = s.subject_id
		INNER JOIN teachers t ON t.id = s.teacher_id
		LEFT JOIN rooms rm ON rm.id = s.room_id
		INNER JOIN academic_years ay ON ay.id = s.academic_year_id
		INNER JOIN semesters sem ON sem.id = s.semester_id
		WHERE s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND sub.deleted_at IS NULL
		  AND t.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		  AND (rm.id IS NULL OR rm.deleted_at IS NULL)
	`

	args := make([]any, 0, 10)
	if search != "" {
		query += " AND (c.name LIKE ? OR sub.code LIKE ? OR sub.name LIKE ? OR t.full_name LIKE ? OR ay.name LIKE ? OR sem.name LIKE ? OR sem.code LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}

	if classID > 0 {
		query += " AND s.class_id = ?"
		args = append(args, classID)
	}
	if teacherID > 0 {
		query += " AND s.teacher_id = ?"
		args = append(args, teacherID)
	}
	if roomID > 0 {
		query += " AND s.room_id = ?"
		args = append(args, roomID)
	}
	if academicYearID > 0 {
		query += " AND s.academic_year_id = ?"
		args = append(args, academicYearID)
	}
	if semesterID > 0 {
		query += " AND s.semester_id = ?"
		args = append(args, semesterID)
	}
	if dayOfWeek != nil {
		query += " AND s.day_of_week = ?"
		args = append(args, *dayOfWeek)
	}

	query += " ORDER BY ay.start_date DESC, sem.id DESC, s.day_of_week ASC, s.start_time ASC, s.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	items := make([]Schedule, 0)
	for rows.Next() {
		item, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schedules: %w", err)
	}

	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Schedule, error) {
	const query = `
		SELECT
			s.id,
			s.class_id,
			c.name,
			s.subject_id,
			sub.code,
			sub.name,
			s.teacher_id,
			t.full_name,
			s.room_id,
			COALESCE(rm.code, ''),
			COALESCE(rm.name, ''),
			s.academic_year_id,
			ay.name,
			s.semester_id,
			sem.code,
			sem.name,
			s.day_of_week,
			TIME_FORMAT(s.start_time, '%H:%i:%s'),
			TIME_FORMAT(s.end_time, '%H:%i:%s'),
			s.created_at,
			s.updated_at,
			s.deleted_at
		FROM schedules s
		INNER JOIN classes c ON c.id = s.class_id
		INNER JOIN subjects sub ON sub.id = s.subject_id
		INNER JOIN teachers t ON t.id = s.teacher_id
		LEFT JOIN rooms rm ON rm.id = s.room_id
		INNER JOIN academic_years ay ON ay.id = s.academic_year_id
		INNER JOIN semesters sem ON sem.id = s.semester_id
		WHERE s.id = ?
		  AND s.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND sub.deleted_at IS NULL
		  AND t.deleted_at IS NULL
		  AND ay.deleted_at IS NULL
		  AND sem.deleted_at IS NULL
		  AND (rm.id IS NULL OR rm.deleted_at IS NULL)
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanSchedule(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get schedule by id: %w", err)
	}

	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Schedule) (*Schedule, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create schedule transaction: %w", err)
	}
	defer tx.Rollback()

	if err := validateReferences(ctx, tx, item, 0); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO schedules (
			class_id,
			subject_id,
			teacher_id,
			room_id,
			academic_year_id,
			semester_id,
			day_of_week,
			start_time,
			end_time
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ClassID, item.SubjectID, item.TeacherID, item.RoomID, item.AcademicYearID, item.SemesterID, item.DayOfWeek, item.StartTime, item.EndTime)
	if err != nil {
		return nil, fmt.Errorf("insert schedule: %w", err)
	}

	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted schedule id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create schedule transaction: %w", err)
	}

	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Schedule) (*Schedule, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update schedule transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureScheduleExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := validateReferences(ctx, tx, item, id); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE schedules
		SET class_id = ?, subject_id = ?, teacher_id = ?, room_id = ?, academic_year_id = ?, semester_id = ?, day_of_week = ?, start_time = ?, end_time = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.ClassID, item.SubjectID, item.TeacherID, item.RoomID, item.AcademicYearID, item.SemesterID, item.DayOfWeek, item.StartTime, item.EndTime, id)
	if err != nil {
		return nil, fmt.Errorf("update schedule: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated schedule affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update schedule transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE schedules
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete schedule: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted schedule affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSchedule(s scanner) (Schedule, error) {
	var item Schedule
	var roomID sql.NullInt64
	var dayOfWeek sql.NullInt64
	var startTime sql.NullString
	var endTime sql.NullString
	err := s.Scan(
		&item.ID,
		&item.ClassID,
		&item.ClassName,
		&item.SubjectID,
		&item.SubjectCode,
		&item.SubjectName,
		&item.TeacherID,
		&item.TeacherFullName,
		&roomID,
		&item.RoomCode,
		&item.RoomName,
		&item.AcademicYearID,
		&item.AcademicYearName,
		&item.SemesterID,
		&item.SemesterCode,
		&item.SemesterName,
		&dayOfWeek,
		&startTime,
		&endTime,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.DeletedAt,
	)
	if err != nil {
		return Schedule{}, err
	}

	if roomID.Valid {
		value := uint64(roomID.Int64)
		item.RoomID = &value
	}
	if dayOfWeek.Valid {
		item.DayOfWeek = uint8(dayOfWeek.Int64)
	}
	if startTime.Valid {
		item.StartTime = startTime.String
	}
	if endTime.Valid {
		item.EndTime = endTime.String
	}

	return item, nil
}

func validateReferences(ctx context.Context, tx *sql.Tx, item Schedule, excludeID uint64) error {
	if err := ensureAcademicYearExists(ctx, tx, item.AcademicYearID); err != nil {
		return err
	}

	if err := ensureSemesterMatchesAcademicYear(ctx, tx, item.SemesterID, item.AcademicYearID); err != nil {
		return err
	}

	classScope, err := ensureClassMatchesAcademicYear(ctx, tx, item.ClassID, item.AcademicYearID)
	if err != nil {
		return err
	}

	if err := ensureSubjectMatchesClassScope(ctx, tx, item.SubjectID, classScope); err != nil {
		return err
	}

	if err := ensureTeacherExists(ctx, tx, item.TeacherID); err != nil {
		return err
	}

	if err := ensureRoomExists(ctx, tx, item.RoomID); err != nil {
		return err
	}

	if err := ensureNoScheduleConflict(ctx, tx, "class_id", item.ClassID, nil, item.AcademicYearID, item.SemesterID, item.DayOfWeek, item.StartTime, item.EndTime, excludeID); err != nil {
		return err
	}
	if err := ensureNoScheduleConflict(ctx, tx, "teacher_id", item.TeacherID, nil, item.AcademicYearID, item.SemesterID, item.DayOfWeek, item.StartTime, item.EndTime, excludeID); err != nil {
		return err
	}
	if item.RoomID != nil {
		if err := ensureNoScheduleConflict(ctx, tx, "room_id", 0, item.RoomID, item.AcademicYearID, item.SemesterID, item.DayOfWeek, item.StartTime, item.EndTime, excludeID); err != nil {
			return err
		}
	}

	return nil
}

func ensureAcademicYearExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM academic_years WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check academic year existence: %w", err)
	}
	if !exists {
		return ErrAcademicYearNotFound
	}
	return nil
}

func ensureSemesterMatchesAcademicYear(ctx context.Context, tx *sql.Tx, semesterID, academicYearID uint64) error {
	var semesterYearID uint64
	err := tx.QueryRowContext(ctx, `
		SELECT academic_year_id
		FROM semesters
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, semesterID).Scan(&semesterYearID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSemesterNotFound
	}
	if err != nil {
		return fmt.Errorf("check semester existence: %w", err)
	}
	if semesterYearID != academicYearID {
		return ErrSemesterAcademicYearMismatch
	}
	return nil
}

func ensureClassMatchesAcademicYear(ctx context.Context, tx *sql.Tx, classID, academicYearID uint64) (classScope, error) {
	var scope classScope
	err := tx.QueryRowContext(ctx, `
		SELECT academic_year_id, department_id, grade_level_id
		FROM classes
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, classID).Scan(&scope.AcademicYearID, &scope.DepartmentID, &scope.GradeLevelID)
	if errors.Is(err, sql.ErrNoRows) {
		return classScope{}, ErrClassNotFound
	}
	if err != nil {
		return classScope{}, fmt.Errorf("check class existence: %w", err)
	}
	if scope.AcademicYearID != academicYearID {
		return classScope{}, ErrClassAcademicYearMismatch
	}
	return scope, nil
}

func ensureSubjectMatchesClassScope(ctx context.Context, tx *sql.Tx, subjectID uint64, classScope classScope) error {
	var scope subjectScope
	err := tx.QueryRowContext(ctx, `
		SELECT department_id, grade_level_id
		FROM subjects
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, subjectID).Scan(&scope.DepartmentID, &scope.GradeLevelID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSubjectNotFound
	}
	if err != nil {
		return fmt.Errorf("check subject existence: %w", err)
	}
	if scope.DepartmentID != classScope.DepartmentID || scope.GradeLevelID != classScope.GradeLevelID {
		return ErrSubjectScopeMismatch
	}
	return nil
}

func ensureTeacherExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM teachers WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check teacher existence: %w", err)
	}
	if !exists {
		return ErrTeacherNotFound
	}
	return nil
}

func ensureRoomExists(ctx context.Context, tx *sql.Tx, id *uint64) error {
	if id == nil {
		return nil
	}

	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM rooms WHERE id = ? AND deleted_at IS NULL
		)
	`, *id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check room existence: %w", err)
	}
	if !exists {
		return ErrRoomNotFound
	}
	return nil
}

func ensureScheduleExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM schedules WHERE id = ? AND deleted_at IS NULL
		)
	`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check schedule existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func ensureNoScheduleConflict(ctx context.Context, tx *sql.Tx, field string, fieldValue uint64, nullableFieldValue *uint64, academicYearID, semesterID uint64, dayOfWeek uint8, startTime, endTime string, excludeID uint64) error {
	if field != "class_id" && field != "teacher_id" && field != "room_id" {
		return fmt.Errorf("unsupported conflict field: %s", field)
	}

	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1
			FROM schedules
			WHERE %s = ?
			  AND academic_year_id = ?
			  AND semester_id = ?
			  AND day_of_week = ?
			  AND deleted_at IS NULL
			  AND start_time < ?
			  AND end_time > ?
			  AND (? = 0 OR id <> ?)
		)
	`, field)

	var key any = fieldValue
	if nullableFieldValue != nil {
		key = *nullableFieldValue
	}

	var exists bool
	err := tx.QueryRowContext(ctx, query, key, academicYearID, semesterID, dayOfWeek, endTime, startTime, excludeID, excludeID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check %s schedule conflict: %w", field, err)
	}
	if !exists {
		return nil
	}

	switch field {
	case "class_id":
		return ErrClassScheduleConflict
	case "teacher_id":
		return ErrTeacherScheduleConflict
	case "room_id":
		return ErrRoomScheduleConflict
	default:
		return nil
	}
}
