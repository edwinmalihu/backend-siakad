package companies

import "time"

type Company struct {
	ID            uint64     `json:"id"`
	CategoryID    *uint64    `json:"category_id,omitempty"`
	CategoryName  string     `json:"category_name,omitempty"`
	Name          string     `json:"name"`
	City          string     `json:"city,omitempty"`
	Address       string     `json:"address,omitempty"`
	ContactPerson string     `json:"contact_person,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	Email         string     `json:"email,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	CategoryID    uint64 `json:"category_id"`
	Name          string `json:"name"`
	City          string `json:"city"`
	Address       string `json:"address"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Status        string `json:"status"`
}

type UpdateRequest struct {
	CategoryID    uint64 `json:"category_id"`
	Name          string `json:"name"`
	City          string `json:"city"`
	Address       string `json:"address"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Status        string `json:"status"`
}
