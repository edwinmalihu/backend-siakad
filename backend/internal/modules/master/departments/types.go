package departments

import "time"

type Department struct {
	ID          uint64     `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	ProgramName string     `json:"program_name,omitempty"`
	FieldName   string     `json:"field_name,omitempty"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ProgramName string `json:"program_name"`
	FieldName   string `json:"field_name"`
	Description string `json:"description"`
}

type UpdateRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	ProgramName string `json:"program_name"`
	FieldName   string `json:"field_name"`
	Description string `json:"description"`
}
