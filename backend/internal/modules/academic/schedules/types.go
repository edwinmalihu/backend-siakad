package schedules

import "time"

type Schedule struct {
	ID               uint64     `json:"id"`
	ClassID          uint64     `json:"class_id"`
	ClassName        string     `json:"class_name"`
	SubjectID        uint64     `json:"subject_id"`
	SubjectCode      string     `json:"subject_code"`
	SubjectName      string     `json:"subject_name"`
	TeacherID        uint64     `json:"teacher_id"`
	TeacherFullName  string     `json:"teacher_full_name"`
	RoomID           *uint64    `json:"room_id,omitempty"`
	RoomCode         string     `json:"room_code,omitempty"`
	RoomName         string     `json:"room_name,omitempty"`
	AcademicYearID   uint64     `json:"academic_year_id"`
	AcademicYearName string     `json:"academic_year_name"`
	SemesterID       uint64     `json:"semester_id"`
	SemesterCode     string     `json:"semester_code"`
	SemesterName     string     `json:"semester_name"`
	DayOfWeek        uint8      `json:"day_of_week"`
	StartTime        string     `json:"start_time"`
	EndTime          string     `json:"end_time"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
}

type CreateRequest struct {
	ClassID        uint64  `json:"class_id"`
	SubjectID      uint64  `json:"subject_id"`
	TeacherID      uint64  `json:"teacher_id"`
	RoomID         *uint64 `json:"room_id"`
	AcademicYearID uint64  `json:"academic_year_id"`
	SemesterID     uint64  `json:"semester_id"`
	DayOfWeek      uint8   `json:"day_of_week"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
}

type UpdateRequest struct {
	ClassID        uint64  `json:"class_id"`
	SubjectID      uint64  `json:"subject_id"`
	TeacherID      uint64  `json:"teacher_id"`
	RoomID         *uint64 `json:"room_id"`
	AcademicYearID uint64  `json:"academic_year_id"`
	SemesterID     uint64  `json:"semester_id"`
	DayOfWeek      uint8   `json:"day_of_week"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
}
