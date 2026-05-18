package students

import "time"

type Student struct {
	ID         uint64     `json:"id"`
	NIS        string     `json:"nis"`
	NISN       string     `json:"nisn,omitempty"`
	FullName   string     `json:"full_name"`
	Gender     string     `json:"gender"`
	BirthPlace string     `json:"birth_place,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	Address    string     `json:"address,omitempty"`
	Phone      string     `json:"phone,omitempty"`
	EntryYear  int        `json:"entry_year"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	NIS        string `json:"nis"`
	NISN       string `json:"nisn"`
	FullName   string `json:"full_name"`
	Gender     string `json:"gender"`
	BirthPlace string `json:"birth_place"`
	BirthDate  string `json:"birth_date"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	EntryYear  int    `json:"entry_year"`
	Status     string `json:"status"`
}

type UpdateRequest struct {
	NIS        string `json:"nis"`
	NISN       string `json:"nisn"`
	FullName   string `json:"full_name"`
	Gender     string `json:"gender"`
	BirthPlace string `json:"birth_place"`
	BirthDate  string `json:"birth_date"`
	Address    string `json:"address"`
	Phone      string `json:"phone"`
	EntryYear  int    `json:"entry_year"`
	Status     string `json:"status"`
}
