package disciplinerecords

import "time"

type DisciplineRecord struct {
	ID                     uint64     `json:"id"`
	StudentID              uint64     `json:"student_id"`
	StudentNIS             string     `json:"student_nis"`
	StudentFullName        string     `json:"student_full_name"`
	DisciplineCategoryID   uint64     `json:"discipline_category_id"`
	DisciplineCategoryName string     `json:"discipline_category_name"`
	Point                  int        `json:"point"`
	RecordedBy             *uint64    `json:"recorded_by,omitempty"`
	RecordedByName         string     `json:"recorded_by_name,omitempty"`
	IncidentDate           time.Time  `json:"incident_date"`
	Description            string     `json:"description,omitempty"`
	ActionTaken            string     `json:"action_taken,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID            uint64 `json:"student_id"`
	DisciplineCategoryID uint64 `json:"discipline_category_id"`
	RecordedBy           uint64 `json:"recorded_by"`
	IncidentDate         string `json:"incident_date"`
	Description          string `json:"description"`
	ActionTaken          string `json:"action_taken"`
}

type UpdateRequest struct {
	StudentID            uint64 `json:"student_id"`
	DisciplineCategoryID uint64 `json:"discipline_category_id"`
	RecordedBy           uint64 `json:"recorded_by"`
	IncidentDate         string `json:"incident_date"`
	Description          string `json:"description"`
	ActionTaken          string `json:"action_taken"`
}
