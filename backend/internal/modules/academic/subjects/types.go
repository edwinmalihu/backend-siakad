package subjects

import "time"

type Subject struct {
	ID             uint64     `json:"id"`
	DepartmentID   uint64     `json:"department_id"`
	DepartmentCode string     `json:"department_code"`
	DepartmentName string     `json:"department_name"`
	GradeLevelID   uint64     `json:"grade_level_id"`
	GradeLevelCode string     `json:"grade_level_code"`
	GradeLevelName string     `json:"grade_level_name"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	SubjectType    string     `json:"subject_type,omitempty"`
	KKM            *float64   `json:"kkm,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	DepartmentID uint64   `json:"department_id"`
	GradeLevelID uint64   `json:"grade_level_id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	SubjectType  string   `json:"subject_type"`
	KKM          *float64 `json:"kkm"`
}

type UpdateRequest struct {
	DepartmentID uint64   `json:"department_id"`
	GradeLevelID uint64   `json:"grade_level_id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	SubjectType  string   `json:"subject_type"`
	KKM          *float64 `json:"kkm"`
}
