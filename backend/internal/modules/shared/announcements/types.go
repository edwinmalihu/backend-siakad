package announcements

import "time"

type Announcement struct {
	ID           uint64     `json:"id"`
	CreatedBy    *uint64    `json:"created_by,omitempty"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	TargetScope  string     `json:"target_scope,omitempty"`
	PublishStart *time.Time `json:"publish_start,omitempty"`
	PublishEnd   *time.Time `json:"publish_end,omitempty"`
	IsPublished  bool       `json:"is_published"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	Title        string `json:"title"`
	Content      string `json:"content"`
	TargetScope  string `json:"target_scope"`
	PublishStart string `json:"publish_start"`
	PublishEnd   string `json:"publish_end"`
	IsPublished  bool   `json:"is_published"`
}

type UpdateRequest struct {
	Title        string `json:"title"`
	Content      string `json:"content"`
	TargetScope  string `json:"target_scope"`
	PublishStart string `json:"publish_start"`
	PublishEnd   string `json:"publish_end"`
	IsPublished  bool   `json:"is_published"`
}
