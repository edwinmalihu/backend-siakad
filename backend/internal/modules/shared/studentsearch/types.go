package studentsearch

type StudentSearchResult struct {
	ID                    uint64 `json:"id"`
	NIS                   string `json:"nis"`
	FullName              string `json:"full_name"`
	Gender                string `json:"gender"`
	Status                string `json:"status"`
	EntryYear             int    `json:"entry_year"`
	ClassName             string `json:"class_name,omitempty"`
	DepartmentCode        string `json:"department_code,omitempty"`
	DepartmentName        string `json:"department_name,omitempty"`
	GradeLevelCode        string `json:"grade_level_code,omitempty"`
	GradeLevelName        string `json:"grade_level_name,omitempty"`
	AcademicYearName      string `json:"academic_year_name,omitempty"`
	SemesterName          string `json:"semester_name,omitempty"`
	EnrollmentStatus      string `json:"enrollment_status,omitempty"`
	DisciplineCount       int    `json:"discipline_count"`
	DisciplinePointTotal  int    `json:"discipline_point_total"`
	AttendanceCount       int    `json:"attendance_count"`
	ExtracurricularCount  int    `json:"extracurricular_count"`
	InternshipStatus      string `json:"internship_status,omitempty"`
	InternshipCompanyName string `json:"internship_company_name,omitempty"`
	AlumniActivity        string `json:"alumni_activity,omitempty"`
	AlumniCompanyName     string `json:"alumni_company_name,omitempty"`
	AlumniCollegeName     string `json:"alumni_college_name,omitempty"`
}

type StudentSearchDetail struct {
	Student           StudentSummary           `json:"student"`
	LatestEnrollment  *EnrollmentSummary       `json:"latest_enrollment,omitempty"`
	LatestMutation    *MutationSummary         `json:"latest_mutation,omitempty"`
	Graduation        *GraduationSummary       `json:"graduation,omitempty"`
	LatestInternship  *InternshipSummary       `json:"latest_internship,omitempty"`
	Alumni            *AlumniSummary           `json:"alumni,omitempty"`
	Stats             StudentSearchStats       `json:"stats"`
	Extracurriculars  []ExtracurricularSummary `json:"extracurriculars"`
	RecentAttendances []AttendanceSummary      `json:"recent_attendances"`
	RecentDisciplines []DisciplineSummary      `json:"recent_disciplines"`
}

type StudentSummary struct {
	ID         uint64 `json:"id"`
	NIS        string `json:"nis"`
	NISN       string `json:"nisn,omitempty"`
	FullName   string `json:"full_name"`
	Gender     string `json:"gender"`
	Status     string `json:"status"`
	EntryYear  int    `json:"entry_year"`
	BirthPlace string `json:"birth_place,omitempty"`
	BirthDate  string `json:"birth_date,omitempty"`
	Address    string `json:"address,omitempty"`
	Phone      string `json:"phone,omitempty"`
}

type EnrollmentSummary struct {
	ClassName        string `json:"class_name"`
	DepartmentCode   string `json:"department_code"`
	DepartmentName   string `json:"department_name"`
	GradeLevelCode   string `json:"grade_level_code"`
	GradeLevelName   string `json:"grade_level_name"`
	AcademicYearName string `json:"academic_year_name"`
	SemesterName     string `json:"semester_name"`
	Status           string `json:"status"`
}

type MutationSummary struct {
	MutationType  string `json:"mutation_type"`
	FromSchool    string `json:"from_school,omitempty"`
	ToSchool      string `json:"to_school,omitempty"`
	Reason        string `json:"reason,omitempty"`
	EffectiveDate string `json:"effective_date,omitempty"`
	Status        string `json:"status"`
}

type GraduationSummary struct {
	AcademicYearName string `json:"academic_year_name,omitempty"`
	GraduationDate   string `json:"graduation_date,omitempty"`
	Status           string `json:"status"`
	Notes            string `json:"notes,omitempty"`
}

type InternshipSummary struct {
	CompanyName      string `json:"company_name"`
	AcademicYearName string `json:"academic_year_name"`
	StartDate        string `json:"start_date,omitempty"`
	EndDate          string `json:"end_date,omitempty"`
	MentorName       string `json:"mentor_name,omitempty"`
	Status           string `json:"status"`
}

type AlumniSummary struct {
	GraduationYear  int    `json:"graduation_year"`
	CurrentActivity string `json:"current_activity,omitempty"`
	CompanyName     string `json:"company_name,omitempty"`
	CollegeName     string `json:"college_name,omitempty"`
	Phone           string `json:"phone,omitempty"`
	Email           string `json:"email,omitempty"`
}

type StudentSearchStats struct {
	AttendanceTotal      int `json:"attendance_total"`
	AttendancePresent    int `json:"attendance_present"`
	AttendanceAbsent     int `json:"attendance_absent"`
	AttendanceExcused    int `json:"attendance_excused"`
	DisciplineCount      int `json:"discipline_count"`
	DisciplinePointTotal int `json:"discipline_point_total"`
	ExtracurricularCount int `json:"extracurricular_count"`
}

type ExtracurricularSummary struct {
	Name             string `json:"name"`
	AcademicYearName string `json:"academic_year_name"`
	Status           string `json:"status"`
}

type AttendanceSummary struct {
	Date   string `json:"date"`
	Status string `json:"status"`
	Class  string `json:"class"`
}

type DisciplineSummary struct {
	IncidentDate string `json:"incident_date"`
	CategoryName string `json:"category_name"`
	Point        int    `json:"point"`
	ActionTaken  string `json:"action_taken,omitempty"`
	Description  string `json:"description,omitempty"`
}
