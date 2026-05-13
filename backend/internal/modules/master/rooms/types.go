package rooms

import "time"

type Room struct {
	ID        uint64     `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Type      string     `json:"type,omitempty"`
	Capacity  *int       `json:"capacity,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Capacity *int   `json:"capacity"`
}

type UpdateRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Capacity *int   `json:"capacity"`
}
