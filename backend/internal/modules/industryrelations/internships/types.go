package internships

import "time"

type Internship struct {
	ID               uint64     `json:"id"`
	StudentID        uint64     `json:"student_id"`
	StudentNIS       string     `json:"student_nis"`
	StudentFullName  string     `json:"student_full_name"`
	CompanyID        uint64     `json:"company_id"`
	CompanyName      string     `json:"company_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	StartDate        *time.Time `json:"start_date,omitempty"`
	EndDate          *time.Time `json:"end_date,omitempty"`
	MentorName       string     `json:"mentor_name,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID      uint64 `json:"student_id"`
	CompanyID      uint64 `json:"company_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	MentorName     string `json:"mentor_name"`
	Status         string `json:"status"`
}

type UpdateRequest struct {
	StudentID      uint64 `json:"student_id"`
	CompanyID      uint64 `json:"company_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	MentorName     string `json:"mentor_name"`
	Status         string `json:"status"`
}
