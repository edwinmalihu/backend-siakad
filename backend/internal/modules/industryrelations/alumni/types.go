package alumni

import "time"

type Alumnus struct {
	ID              uint64     `json:"id"`
	StudentID       uint64     `json:"student_id"`
	StudentNIS      string     `json:"student_nis"`
	StudentFullName string     `json:"student_full_name"`
	GraduationYear  int        `json:"graduation_year"`
	CurrentActivity string     `json:"current_activity,omitempty"`
	CompanyName     string     `json:"company_name,omitempty"`
	CollegeName     string     `json:"college_name,omitempty"`
	Phone           string     `json:"phone,omitempty"`
	Email           string     `json:"email,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID       uint64 `json:"student_id"`
	GraduationYear  int    `json:"graduation_year"`
	CurrentActivity string `json:"current_activity"`
	CompanyName     string `json:"company_name"`
	CollegeName     string `json:"college_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
}

type UpdateRequest struct {
	StudentID       uint64 `json:"student_id"`
	GraduationYear  int    `json:"graduation_year"`
	CurrentActivity string `json:"current_activity"`
	CompanyName     string `json:"company_name"`
	CollegeName     string `json:"college_name"`
	Phone           string `json:"phone"`
	Email           string `json:"email"`
}
