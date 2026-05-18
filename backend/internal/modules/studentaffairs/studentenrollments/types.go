package studentenrollments

import "time"

type StudentEnrollment struct {
	ID               uint64     `json:"id"`
	StudentID        uint64     `json:"student_id"`
	StudentNIS       string     `json:"student_nis"`
	StudentFullName  string     `json:"student_full_name"`
	ClassID          uint64     `json:"class_id"`
	ClassName        string     `json:"class_name"`
	DepartmentCode   string     `json:"department_code"`
	DepartmentName   string     `json:"department_name"`
	GradeLevelCode   string     `json:"grade_level_code"`
	GradeLevelName   string     `json:"grade_level_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	SemesterID       uint64     `json:"semester_id"`
	SemesterCode     string     `json:"semester_code"`
	SemesterName     string     `json:"semester_name"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID      uint64 `json:"student_id"`
	ClassID        uint64 `json:"class_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	SemesterID     uint64 `json:"semester_id"`
	Status         string `json:"status"`
}

type UpdateRequest struct {
	StudentID      uint64 `json:"student_id"`
	ClassID        uint64 `json:"class_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	SemesterID     uint64 `json:"semester_id"`
	Status         string `json:"status"`
}
