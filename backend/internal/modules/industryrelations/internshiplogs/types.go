package internshiplogs

import (
	"time"
)

type InternshipLog struct {
	ID             uint64     `json:"id"`
	InternshipID   uint64     `json:"internship_id"`
	StudentName    string     `json:"student_name"`
	CompanyName    string     `json:"company_name"`
	LogDate        time.Time  `json:"log_date"`
	Activity       string     `json:"activity"`
	Notes          string     `json:"notes,omitempty"`
	SupervisorName string     `json:"supervisor_name,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	InternshipID   uint64 `json:"internship_id"`
	LogDate        string `json:"log_date"`
	Activity       string `json:"activity"`
	Notes          string `json:"notes"`
	SupervisorName string `json:"supervisor_name"`
}

type UpdateRequest struct {
	InternshipID   uint64 `json:"internship_id"`
	LogDate        string `json:"log_date"`
	Activity       string `json:"activity"`
	Notes          string `json:"notes"`
	SupervisorName string `json:"supervisor_name"`
}
