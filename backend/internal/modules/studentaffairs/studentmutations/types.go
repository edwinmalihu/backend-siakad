package studentmutations

import "time"

type StudentMutation struct {
	ID               uint64     `json:"id"`
	StudentID        uint64     `json:"student_id"`
	StudentNIS       string     `json:"student_nis"`
	StudentFullName  string     `json:"student_full_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	SemesterID       uint64     `json:"semester_id"`
	SemesterCode     string     `json:"semester_code"`
	SemesterName     string     `json:"semester_name"`
	MutationType     string     `json:"mutation_type"`
	FromSchool       string     `json:"from_school,omitempty"`
	ToSchool         string     `json:"to_school,omitempty"`
	Reason           string     `json:"reason,omitempty"`
	EffectiveDate    *time.Time `json:"effective_date,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID      uint64 `json:"student_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	SemesterID     uint64 `json:"semester_id"`
	MutationType   string `json:"mutation_type"`
	FromSchool     string `json:"from_school"`
	ToSchool       string `json:"to_school"`
	Reason         string `json:"reason"`
	EffectiveDate  string `json:"effective_date"`
	Status         string `json:"status"`
}

type UpdateRequest struct {
	StudentID      uint64 `json:"student_id"`
	AcademicYearID uint64 `json:"academic_year_id"`
	SemesterID     uint64 `json:"semester_id"`
	MutationType   string `json:"mutation_type"`
	FromSchool     string `json:"from_school"`
	ToSchool       string `json:"to_school"`
	Reason         string `json:"reason"`
	EffectiveDate  string `json:"effective_date"`
	Status         string `json:"status"`
}
