package semesters

import "time"

type Semester struct {
	ID             uint64     `json:"id"`
	AcademicYearID uint64     `json:"academic_year_id"`
	AcademicYear   string     `json:"academic_year"`
	Name           string     `json:"name"`
	Code           string     `json:"code"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	AcademicYearID uint64 `json:"academic_year_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	IsActive       bool   `json:"is_active"`
}

type UpdateRequest struct {
	AcademicYearID uint64 `json:"academic_year_id"`
	Name           string `json:"name"`
	Code           string `json:"code"`
	IsActive       bool   `json:"is_active"`
}
