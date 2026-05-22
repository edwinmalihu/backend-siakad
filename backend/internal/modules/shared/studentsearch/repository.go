package studentsearch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("student not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Search(ctx context.Context, search, status string) ([]StudentSearchResult, error) {
	query := `
		SELECT
			s.id,
			s.nis,
			s.full_name,
			s.gender,
			s.status,
			s.entry_year,
			COALESCE(le.class_name, ''),
			COALESCE(le.department_code, ''),
			COALESCE(le.department_name, ''),
			COALESCE(le.grade_level_code, ''),
			COALESCE(le.grade_level_name, ''),
			COALESCE(le.academic_year_name, ''),
			COALESCE(le.semester_name, ''),
			COALESCE(le.enrollment_status, ''),
			COALESCE(ds.discipline_count, 0),
			COALESCE(ds.discipline_point_total, 0),
			COALESCE(att.attendance_count, 0),
			COALESCE(ex.extracurricular_count, 0),
			COALESCE(li.internship_status, ''),
			COALESCE(li.company_name, ''),
			COALESCE(al.current_activity, ''),
			COALESCE(al.company_name, ''),
			COALESCE(al.college_name, '')
		FROM students s
		LEFT JOIN (
			SELECT
				se.student_id,
				c.name AS class_name,
				d.code AS department_code,
				d.name AS department_name,
				gl.code AS grade_level_code,
				gl.name AS grade_level_name,
				ay.name AS academic_year_name,
				sem.name AS semester_name,
				se.status AS enrollment_status
			FROM student_enrollments se
			INNER JOIN (
				SELECT student_id, MAX(id) AS max_id
				FROM student_enrollments
				WHERE deleted_at IS NULL
				GROUP BY student_id
			) latest ON latest.max_id = se.id
			INNER JOIN classes c ON c.id = se.class_id AND c.deleted_at IS NULL
			INNER JOIN departments d ON d.id = c.department_id AND d.deleted_at IS NULL
			INNER JOIN grade_levels gl ON gl.id = c.grade_level_id AND gl.deleted_at IS NULL
			INNER JOIN academic_years ay ON ay.id = se.academic_year_id AND ay.deleted_at IS NULL
			INNER JOIN semesters sem ON sem.id = se.semester_id AND sem.deleted_at IS NULL
			WHERE se.deleted_at IS NULL
		) le ON le.student_id = s.id
		LEFT JOIN (
			SELECT
				dr.student_id,
				COUNT(*) AS discipline_count,
				COALESCE(SUM(dc.point), 0) AS discipline_point_total
			FROM discipline_records dr
			INNER JOIN discipline_categories dc ON dc.id = dr.discipline_category_id AND dc.deleted_at IS NULL
			WHERE dr.deleted_at IS NULL
			GROUP BY dr.student_id
		) ds ON ds.student_id = s.id
		LEFT JOIN (
			SELECT student_id, COUNT(*) AS attendance_count
			FROM attendances
			WHERE deleted_at IS NULL
			GROUP BY student_id
		) att ON att.student_id = s.id
		LEFT JOIN (
			SELECT student_id, COUNT(*) AS extracurricular_count
			FROM extracurricular_members
			WHERE deleted_at IS NULL
			GROUP BY student_id
		) ex ON ex.student_id = s.id
		LEFT JOIN (
			SELECT
				i.student_id,
				i.status AS internship_status,
				c.name AS company_name
			FROM internships i
			INNER JOIN (
				SELECT student_id, MAX(id) AS max_id
				FROM internships
				WHERE deleted_at IS NULL
				GROUP BY student_id
			) latest ON latest.max_id = i.id
			INNER JOIN companies c ON c.id = i.company_id AND c.deleted_at IS NULL
			WHERE i.deleted_at IS NULL
		) li ON li.student_id = s.id
		LEFT JOIN (
			SELECT
				a.student_id,
				a.current_activity,
				a.company_name,
				a.college_name
			FROM alumni a
			WHERE a.deleted_at IS NULL
		) al ON al.student_id = s.id
		WHERE s.deleted_at IS NULL
	`
	args := make([]any, 0, 3)
	if search != "" {
		query += " AND (s.nis LIKE ? OR s.nisn LIKE ? OR s.full_name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if status != "" {
		query += " AND s.status = ?"
		args = append(args, status)
	}
	query += " ORDER BY s.full_name ASC, s.id DESC LIMIT 50"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query student search results: %w", err)
	}
	defer rows.Close()

	items := make([]StudentSearchResult, 0)
	for rows.Next() {
		var item StudentSearchResult
		if err := rows.Scan(
			&item.ID,
			&item.NIS,
			&item.FullName,
			&item.Gender,
			&item.Status,
			&item.EntryYear,
			&item.ClassName,
			&item.DepartmentCode,
			&item.DepartmentName,
			&item.GradeLevelCode,
			&item.GradeLevelName,
			&item.AcademicYearName,
			&item.SemesterName,
			&item.EnrollmentStatus,
			&item.DisciplineCount,
			&item.DisciplinePointTotal,
			&item.AttendanceCount,
			&item.ExtracurricularCount,
			&item.InternshipStatus,
			&item.InternshipCompanyName,
			&item.AlumniActivity,
			&item.AlumniCompanyName,
			&item.AlumniCollegeName,
		); err != nil {
			return nil, fmt.Errorf("scan student search result: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate student search results: %w", err)
	}
	return items, nil
}

func (r *Repository) GetDetail(ctx context.Context, studentID uint64) (*StudentSearchDetail, error) {
	student, err := r.getStudentSummary(ctx, studentID)
	if err != nil {
		return nil, err
	}

	latestEnrollment, err := r.getLatestEnrollment(ctx, studentID)
	if err != nil {
		return nil, err
	}
	latestMutation, err := r.getLatestMutation(ctx, studentID)
	if err != nil {
		return nil, err
	}
	graduation, err := r.getGraduation(ctx, studentID)
	if err != nil {
		return nil, err
	}
	latestInternship, err := r.getLatestInternship(ctx, studentID)
	if err != nil {
		return nil, err
	}
	alumni, err := r.getAlumni(ctx, studentID)
	if err != nil {
		return nil, err
	}
	stats, err := r.getStats(ctx, studentID)
	if err != nil {
		return nil, err
	}
	extracurriculars, err := r.getExtracurriculars(ctx, studentID)
	if err != nil {
		return nil, err
	}
	recentAttendances, err := r.getRecentAttendances(ctx, studentID)
	if err != nil {
		return nil, err
	}
	recentDisciplines, err := r.getRecentDisciplines(ctx, studentID)
	if err != nil {
		return nil, err
	}

	return &StudentSearchDetail{
		Student:           *student,
		LatestEnrollment:  latestEnrollment,
		LatestMutation:    latestMutation,
		Graduation:        graduation,
		LatestInternship:  latestInternship,
		Alumni:            alumni,
		Stats:             *stats,
		Extracurriculars:  extracurriculars,
		RecentAttendances: recentAttendances,
		RecentDisciplines: recentDisciplines,
	}, nil
}

func (r *Repository) getStudentSummary(ctx context.Context, studentID uint64) (*StudentSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, nis, nisn, full_name, gender, status, entry_year, birth_place, birth_date, address, phone
		FROM students
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, studentID)
	var item StudentSummary
	var nisn sql.NullString
	var birthPlace sql.NullString
	var birthDate sql.NullTime
	var address sql.NullString
	var phone sql.NullString
	if err := row.Scan(&item.ID, &item.NIS, &nisn, &item.FullName, &item.Gender, &item.Status, &item.EntryYear, &birthPlace, &birthDate, &address, &phone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get student summary: %w", err)
	}
	if nisn.Valid {
		item.NISN = nisn.String
	}
	if birthPlace.Valid {
		item.BirthPlace = birthPlace.String
	}
	if birthDate.Valid {
		item.BirthDate = birthDate.Time.Format("2006-01-02")
	}
	if address.Valid {
		item.Address = address.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	return &item, nil
}

func (r *Repository) getLatestEnrollment(ctx context.Context, studentID uint64) (*EnrollmentSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			c.name,
			d.code,
			d.name,
			gl.code,
			gl.name,
			ay.name,
			sem.name,
			se.status
		FROM student_enrollments se
		INNER JOIN classes c ON c.id = se.class_id AND c.deleted_at IS NULL
		INNER JOIN departments d ON d.id = c.department_id AND d.deleted_at IS NULL
		INNER JOIN grade_levels gl ON gl.id = c.grade_level_id AND gl.deleted_at IS NULL
		INNER JOIN academic_years ay ON ay.id = se.academic_year_id AND ay.deleted_at IS NULL
		INNER JOIN semesters sem ON sem.id = se.semester_id AND sem.deleted_at IS NULL
		WHERE se.student_id = ? AND se.deleted_at IS NULL
		ORDER BY se.id DESC
		LIMIT 1
	`, studentID)
	var item EnrollmentSummary
	if err := row.Scan(&item.ClassName, &item.DepartmentCode, &item.DepartmentName, &item.GradeLevelCode, &item.GradeLevelName, &item.AcademicYearName, &item.SemesterName, &item.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest enrollment: %w", err)
	}
	return &item, nil
}

func (r *Repository) getLatestMutation(ctx context.Context, studentID uint64) (*MutationSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT mutation_type, from_school, to_school, reason, effective_date, status
		FROM student_mutations
		WHERE student_id = ? AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 1
	`, studentID)
	var item MutationSummary
	var fromSchool, toSchool, reason sql.NullString
	var effectiveDate sql.NullTime
	if err := row.Scan(&item.MutationType, &fromSchool, &toSchool, &reason, &effectiveDate, &item.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest mutation: %w", err)
	}
	if fromSchool.Valid {
		item.FromSchool = fromSchool.String
	}
	if toSchool.Valid {
		item.ToSchool = toSchool.String
	}
	if reason.Valid {
		item.Reason = reason.String
	}
	if effectiveDate.Valid {
		item.EffectiveDate = effectiveDate.Time.Format("2006-01-02")
	}
	return &item, nil
}

func (r *Repository) getGraduation(ctx context.Context, studentID uint64) (*GraduationSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(ay.name, ''),
			sg.graduation_date,
			sg.status,
			sg.notes
		FROM student_graduations sg
		LEFT JOIN academic_years ay ON ay.id = sg.academic_year_id AND ay.deleted_at IS NULL
		WHERE sg.student_id = ? AND sg.deleted_at IS NULL
		ORDER BY sg.id DESC
		LIMIT 1
	`, studentID)
	var item GraduationSummary
	var graduationDate sql.NullTime
	var notes sql.NullString
	if err := row.Scan(&item.AcademicYearName, &graduationDate, &item.Status, &notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get graduation summary: %w", err)
	}
	if graduationDate.Valid {
		item.GraduationDate = graduationDate.Time.Format("2006-01-02")
	}
	if notes.Valid {
		item.Notes = notes.String
	}
	return &item, nil
}

func (r *Repository) getLatestInternship(ctx context.Context, studentID uint64) (*InternshipSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT c.name, ay.name, i.start_date, i.end_date, i.mentor_name, i.status
		FROM internships i
		INNER JOIN companies c ON c.id = i.company_id AND c.deleted_at IS NULL
		INNER JOIN academic_years ay ON ay.id = i.academic_year_id AND ay.deleted_at IS NULL
		WHERE i.student_id = ? AND i.deleted_at IS NULL
		ORDER BY i.id DESC
		LIMIT 1
	`, studentID)
	var item InternshipSummary
	var startDate, endDate sql.NullTime
	var mentorName sql.NullString
	if err := row.Scan(&item.CompanyName, &item.AcademicYearName, &startDate, &endDate, &mentorName, &item.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest internship: %w", err)
	}
	if startDate.Valid {
		item.StartDate = startDate.Time.Format("2006-01-02")
	}
	if endDate.Valid {
		item.EndDate = endDate.Time.Format("2006-01-02")
	}
	if mentorName.Valid {
		item.MentorName = mentorName.String
	}
	return &item, nil
}

func (r *Repository) getAlumni(ctx context.Context, studentID uint64) (*AlumniSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT graduation_year, current_activity, company_name, college_name, phone, email
		FROM alumni
		WHERE student_id = ? AND deleted_at IS NULL
		LIMIT 1
	`, studentID)
	var item AlumniSummary
	var currentActivity, companyName, collegeName, phone, email sql.NullString
	if err := row.Scan(&item.GraduationYear, &currentActivity, &companyName, &collegeName, &phone, &email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get alumni summary: %w", err)
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
	return &item, nil
}

func (r *Repository) getStats(ctx context.Context, studentID uint64) (*StudentSearchStats, error) {
	var stats StudentSearchStats
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM attendances
		WHERE student_id = ? AND deleted_at IS NULL
		GROUP BY status
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("query attendance stats: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scan attendance stats: %w", err)
		}
		stats.AttendanceTotal += count
		switch strings.ToLower(status) {
		case "present", "hadir":
			stats.AttendancePresent += count
		case "absent", "alpha":
			stats.AttendanceAbsent += count
		default:
			stats.AttendanceExcused += count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attendance stats: %w", err)
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(dc.point), 0)
		FROM discipline_records dr
		INNER JOIN discipline_categories dc ON dc.id = dr.discipline_category_id AND dc.deleted_at IS NULL
		WHERE dr.student_id = ? AND dr.deleted_at IS NULL
	`, studentID)
	if err := row.Scan(&stats.DisciplineCount, &stats.DisciplinePointTotal); err != nil {
		return nil, fmt.Errorf("get discipline stats: %w", err)
	}

	row = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM extracurricular_members
		WHERE student_id = ? AND deleted_at IS NULL
	`, studentID)
	if err := row.Scan(&stats.ExtracurricularCount); err != nil {
		return nil, fmt.Errorf("get extracurricular stats: %w", err)
	}

	return &stats, nil
}

func (r *Repository) getExtracurriculars(ctx context.Context, studentID uint64) ([]ExtracurricularSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT e.name, ay.name, em.status
		FROM extracurricular_members em
		INNER JOIN extracurriculars e ON e.id = em.extracurricular_id AND e.deleted_at IS NULL
		INNER JOIN academic_years ay ON ay.id = em.academic_year_id AND ay.deleted_at IS NULL
		WHERE em.student_id = ? AND em.deleted_at IS NULL
		ORDER BY em.id DESC
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("query extracurricular summaries: %w", err)
	}
	defer rows.Close()
	items := make([]ExtracurricularSummary, 0)
	for rows.Next() {
		var item ExtracurricularSummary
		if err := rows.Scan(&item.Name, &item.AcademicYearName, &item.Status); err != nil {
			return nil, fmt.Errorf("scan extracurricular summary: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate extracurricular summaries: %w", err)
	}
	return items, nil
}

func (r *Repository) getRecentAttendances(ctx context.Context, studentID uint64) ([]AttendanceSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.attendance_date, a.status, c.name
		FROM attendances a
		INNER JOIN classes c ON c.id = a.class_id AND c.deleted_at IS NULL
		WHERE a.student_id = ? AND a.deleted_at IS NULL
		ORDER BY a.attendance_date DESC, a.id DESC
		LIMIT 5
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("query recent attendances: %w", err)
	}
	defer rows.Close()
	items := make([]AttendanceSummary, 0)
	for rows.Next() {
		var item AttendanceSummary
		var date time.Time
		if err := rows.Scan(&date, &item.Status, &item.Class); err != nil {
			return nil, fmt.Errorf("scan recent attendance: %w", err)
		}
		item.Date = date.Format("2006-01-02")
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent attendances: %w", err)
	}
	return items, nil
}

func (r *Repository) getRecentDisciplines(ctx context.Context, studentID uint64) ([]DisciplineSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT dr.incident_date, dc.name, dc.point, dr.action_taken, dr.description
		FROM discipline_records dr
		INNER JOIN discipline_categories dc ON dc.id = dr.discipline_category_id AND dc.deleted_at IS NULL
		WHERE dr.student_id = ? AND dr.deleted_at IS NULL
		ORDER BY dr.incident_date DESC, dr.id DESC
		LIMIT 5
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("query recent discipline records: %w", err)
	}
	defer rows.Close()
	items := make([]DisciplineSummary, 0)
	for rows.Next() {
		var item DisciplineSummary
		var incidentDate time.Time
		var actionTaken, description sql.NullString
		if err := rows.Scan(&incidentDate, &item.CategoryName, &item.Point, &actionTaken, &description); err != nil {
			return nil, fmt.Errorf("scan recent discipline: %w", err)
		}
		item.IncidentDate = incidentDate.Format("2006-01-02")
		if actionTaken.Valid {
			item.ActionTaken = actionTaken.String
		}
		if description.Valid {
			item.Description = description.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent discipline records: %w", err)
	}
	return items, nil
}
