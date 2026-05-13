package teachers

import "time"

type Teacher struct {
	ID               uint64     `json:"id"`
	NIP              string     `json:"nip,omitempty"`
	NUPTK            string     `json:"nuptk,omitempty"`
	FullName         string     `json:"full_name"`
	Gender           string     `json:"gender,omitempty"`
	Address          string     `json:"address,omitempty"`
	Phone            string     `json:"phone,omitempty"`
	Email            string     `json:"email,omitempty"`
	EmploymentStatus string     `json:"employment_status,omitempty"`
	Position         string     `json:"position,omitempty"`
	PhotoURL         string     `json:"photo_url,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	NIP              string `json:"nip"`
	NUPTK            string `json:"nuptk"`
	FullName         string `json:"full_name"`
	Gender           string `json:"gender"`
	Address          string `json:"address"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	EmploymentStatus string `json:"employment_status"`
	Position         string `json:"position"`
	PhotoURL         string `json:"photo_url"`
	Status           string `json:"status"`
}

type UpdateRequest struct {
	NIP              string `json:"nip"`
	NUPTK            string `json:"nuptk"`
	FullName         string `json:"full_name"`
	Gender           string `json:"gender"`
	Address          string `json:"address"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	EmploymentStatus string `json:"employment_status"`
	Position         string `json:"position"`
	PhotoURL         string `json:"photo_url"`
	Status           string `json:"status"`
}
