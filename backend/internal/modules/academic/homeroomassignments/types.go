package homeroomassignments

import "time"

type HomeroomAssignment struct {
	ID               uint64     `json:"id"`
	TeacherID        uint64     `json:"teacher_id"`
	TeacherNIP       string     `json:"teacher_nip,omitempty"`
	TeacherFullName  string     `json:"teacher_full_name"`
	ClassID          uint64     `json:"class_id"`
	ClassName        string     `json:"class_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	SemesterID       uint64     `json:"semester_id"`
	SemesterCode     string     `json:"semester_code"`
	SemesterName     string     `json:"semester_name"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	TeacherID      uint64 `json:"teacher_id"`
	ClassID        uint64 `json:"class_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	SemesterID     uint64 `json:"semester_id"`
}

type UpdateRequest struct {
	TeacherID      uint64 `json:"teacher_id"`
	ClassID        uint64 `json:"class_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	SemesterID     uint64 `json:"semester_id"`
}
