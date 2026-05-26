package assessmentcomponents

import "time"

type AssessmentComponent struct {
	ID               uint64     `json:"id"`
	SubjectID        uint64     `json:"subject_id"`
	SubjectCode      string     `json:"subject_code"`
	SubjectName      string     `json:"subject_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	SemesterID       uint64     `json:"semester_id"`
	SemesterCode     string     `json:"semester_code"`
	SemesterName     string     `json:"semester_name"`
	Name             string     `json:"name"`
	Weight           float64    `json:"weight"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	SubjectID      uint64  `json:"subject_id"`
	AcademicYearID uint64  `json:"academic_year_id"`
	SemesterID     uint64  `json:"semester_id"`
	Name           string  `json:"name"`
	Weight         float64 `json:"weight"`
}

type UpdateRequest struct {
	SubjectID      uint64  `json:"subject_id"`
	AcademicYearID uint64  `json:"academic_year_id"`
	SemesterID     uint64  `json:"semester_id"`
	Name           string  `json:"name"`
	Weight         float64 `json:"weight"`
}
