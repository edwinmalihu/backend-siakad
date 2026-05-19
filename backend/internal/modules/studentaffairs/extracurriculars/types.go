package extracurriculars

import "time"

type Extracurricular struct {
	ID               uint64     `json:"id"`
	CoachTeacherID   *uint64    `json:"coach_teacher_id,omitempty"`
	CoachTeacherName string     `json:"coach_teacher_name,omitempty"`
	Name             string     `json:"name"`
	Description      string     `json:"description,omitempty"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	CoachTeacherID uint64 `json:"coach_teacher_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IsActive       bool   `json:"is_active"`
}

type UpdateRequest struct {
	CoachTeacherID uint64 `json:"coach_teacher_id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	IsActive       bool   `json:"is_active"`
}
