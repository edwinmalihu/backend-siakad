package industrycategories

import "time"

type IndustryCategory struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
