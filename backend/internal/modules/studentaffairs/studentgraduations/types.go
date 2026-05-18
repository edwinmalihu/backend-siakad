package studentgraduations

import "time"

type StudentGraduation struct {
	ID               uint64     `json:"id"`
	StudentID        uint64     `json:"student_id"`
	StudentNIS       string     `json:"student_nis"`
	StudentFullName  string     `json:"student_full_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	GraduationDate   *time.Time `json:"graduation_date,omitempty"`
	Status           string     `json:"status"`
	Notes            string     `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID      uint64 `json:"student_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	GraduationDate string `json:"graduation_date"`
	Status         string `json:"status"`
	Notes          string `json:"notes"`
}

type UpdateRequest struct {
	StudentID      uint64 `json:"student_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	GraduationDate string `json:"graduation_date"`
	Status         string `json:"status"`
	Notes          string `json:"notes"`
}
