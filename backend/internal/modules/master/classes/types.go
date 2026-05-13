package classes

import "time"

type Class struct {
	ID               uint64     `json:"id"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	DepartmentID     uint64     `json:"department_id"`
	DepartmentCode   string     `json:"department_code"`
	DepartmentName   string     `json:"department_name"`
	GradeLevelID     uint64     `json:"grade_level_id"`
	GradeLevelCode   string     `json:"grade_level_code"`
	GradeLevelName   string     `json:"grade_level_name"`
	Name             string     `json:"name"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	AcademicYearID uint64 `json:"academic_year_id"`
	DepartmentID   uint64 `json:"department_id"`
	GradeLevelID   uint64 `json:"grade_level_id"`
	Name           string `json:"name"`
	IsActive       bool   `json:"is_active"`
}

type UpdateRequest struct {
	AcademicYearID uint64 `json:"academic_year_id"`
	DepartmentID   uint64 `json:"department_id"`
	GradeLevelID   uint64 `json:"grade_level_id"`
	Name           string `json:"name"`
	IsActive       bool   `json:"is_active"`
}
