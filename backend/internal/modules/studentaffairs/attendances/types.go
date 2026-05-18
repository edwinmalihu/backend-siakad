package attendances

import "time"

type Attendance struct {
	ID               uint64     `json:"id"`
	StudentID        uint64     `json:"student_id"`
	StudentNIS       string     `json:"student_nis"`
	StudentFullName  string     `json:"student_full_name"`
	ClassID          uint64     `json:"class_id"`
	ClassName        string     `json:"class_name"`
	DepartmentCode   string     `json:"department_code"`
	DepartmentName   string     `json:"department_name"`
	GradeLevelCode   string     `json:"grade_level_code"`
	GradeLevelName   string     `json:"grade_level_name"`
	AttendanceDate   time.Time  `json:"attendance_date"`
	Status           string     `json:"status"`
	Notes            string     `json:"notes,omitempty"`
	RecordedBy       *uint64    `json:"recorded_by,omitempty"`
	RecordedByName   string     `json:"recorded_by_name,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	StudentID      uint64 `json:"student_id"`
	ClassID        uint64 `json:"class_id"`
	AttendanceDate string `json:"attendance_date"`
	Status         string `json:"status"`
	Notes          string `json:"notes"`
	RecordedBy     uint64 `json:"recorded_by"`
}

type UpdateRequest struct {
	StudentID      uint64 `json:"student_id"`
	ClassID        uint64 `json:"class_id"`
	AttendanceDate string `json:"attendance_date"`
	Status         string `json:"status"`
	Notes          string `json:"notes"`
	RecordedBy     uint64 `json:"recorded_by"`
}
