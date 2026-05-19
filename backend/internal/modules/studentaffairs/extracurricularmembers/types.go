package extracurricularmembers

import "time"

type ExtracurricularMember struct {
	ID                  uint64     `json:"id"`
	ExtracurricularID   uint64     `json:"extracurricular_id"`
	ExtracurricularName string     `json:"extracurricular_name"`
	StudentID           uint64     `json:"student_id"`
	StudentNIS          string     `json:"student_nis"`
	StudentFullName     string     `json:"student_full_name"`
	AcademicYearID      uint64     `json:"academic_year_id"`
	AcademicYearName    string     `json:"academic_year_name"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	ExtracurricularID uint64 `json:"extracurricular_id"`
	StudentID         uint64 `json:"student_id"`
	AcademicYearID    uint64 `json:"academic_year_id"`
	Status            string `json:"status"`
}

type UpdateRequest struct {
	ExtracurricularID uint64 `json:"extracurricular_id"`
	StudentID         uint64 `json:"student_id"`
	AcademicYearID    uint64 `json:"academic_year_id"`
	Status            string `json:"status"`
}
