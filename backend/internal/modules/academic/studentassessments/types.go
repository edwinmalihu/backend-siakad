package studentassessments

import "time"

type StudentAssessment struct {
	ID                   uint64     `json:"id"`
	StudentID            uint64     `json:"student_id"`
	StudentName          string     `json:"student_name"`
	StudentNIS           string     `json:"student_nis"`
	ClassID              uint64     `json:"class_id"`
	ClassName            string     `json:"class_name"`
	SubjectID            uint64     `json:"subject_id"`
	SubjectCode          string     `json:"subject_code"`
	SubjectName          string     `json:"subject_name"`
	AssessmentComponentID uint64   `json:"assessment_component_id"`
	ComponentName        string     `json:"component_name"`
	ComponentWeight      float64    `json:"component_weight"`
	TeacherID            uint64     `json:"teacher_id"`
	TeacherName          string     `json:"teacher_name"`
	Score                float64    `json:"score"`
	AcademicYearID       uint64     `json:"academic_year_id"`
	AcademicYearName     string     `json:"academic_year_name"`
	SemesterID           uint64     `json:"semester_id"`
	SemesterCode         string     `json:"semester_code"`
	SemesterName         string     `json:"semester_name"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID            uint64  `json:"student_id"`
	ClassID              uint64  `json:"class_id"`
	SubjectID            uint64  `json:"subject_id"`
	AssessmentComponentID uint64 `json:"assessment_component_id"`
	TeacherID            uint64  `json:"teacher_id"`
	Score                float64 `json:"score"`
	AcademicYearID       uint64  `json:"academic_year_id"`
	SemesterID           uint64  `json:"semester_id"`
}

type UpdateRequest struct {
	StudentID            uint64  `json:"student_id"`
	ClassID              uint64  `json:"class_id"`
	SubjectID            uint64  `json:"subject_id"`
	AssessmentComponentID uint64 `json:"assessment_component_id"`
	TeacherID            uint64  `json:"teacher_id"`
	Score                float64 `json:"score"`
	AcademicYearID       uint64  `json:"academic_year_id"`
	SemesterID           uint64  `json:"semester_id"`
}
