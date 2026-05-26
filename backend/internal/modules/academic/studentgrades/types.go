package studentgrades

import "time"

type StudentGrade struct {
	ID               uint64     `json:"id"`
	StudentID        uint64     `json:"student_id"`
	StudentName      string     `json:"student_name"`
	StudentNIS       string     `json:"student_nis"`
	ClassID          uint64     `json:"class_id"`
	ClassName        string     `json:"class_name"`
	SubjectID        uint64     `json:"subject_id"`
	SubjectCode      string     `json:"subject_code"`
	SubjectName      string     `json:"subject_name"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	SemesterID       uint64     `json:"semester_id"`
	SemesterCode     string     `json:"semester_code"`
	SemesterName     string     `json:"semester_name"`
	FinalScore       float64    `json:"final_score"`
	GradeLetter      string     `json:"grade_letter,omitempty"`
	Predicate        string     `json:"predicate,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID      uint64  `json:"student_id"`
	ClassID        uint64  `json:"class_id"`
	SubjectID      uint64  `json:"subject_id"`
	AcademicYearID uint64  `json:"academic_year_id"`
	SemesterID     uint64  `json:"semester_id"`
	FinalScore     float64 `json:"final_score"`
	GradeLetter    string  `json:"grade_letter"`
	Predicate      string  `json:"predicate"`
}

type UpdateRequest struct {
	StudentID      uint64  `json:"student_id"`
	ClassID        uint64  `json:"class_id"`
	SubjectID      uint64  `json:"subject_id"`
	AcademicYearID uint64  `json:"academic_year_id"`
	SemesterID     uint64  `json:"semester_id"`
	FinalScore     float64 `json:"final_score"`
	GradeLetter    string  `json:"grade_letter"`
	Predicate      string  `json:"predicate"`
}
