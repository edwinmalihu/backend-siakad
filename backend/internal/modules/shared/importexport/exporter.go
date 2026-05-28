package importexport

import (
	"context"
	"database/sql"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportFilters holds optional query parameters for export
type ExportFilters struct {
	Status    string // filter by status (active/inactive)
	Gender    string // filter by gender (male/female)
	EntryYear int    // filter by entry_year (0 = no filter)
	IsActive  *int   // filter by is_active (nil = no filter)
	Search    string // free-text search on name/code
}

func parseExportFilters(params url.Values) ExportFilters {
	f := ExportFilters{}

	if v := strings.TrimSpace(params.Get("status")); v != "" {
		f.Status = strings.ToLower(v)
	}
	if v := strings.TrimSpace(params.Get("gender")); v != "" {
		f.Gender = strings.ToLower(v)
	}
	if v := strings.TrimSpace(params.Get("entry_year")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.EntryYear = n
		}
	}
	if v := strings.TrimSpace(params.Get("is_active")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && (n == 0 || n == 1) {
			f.IsActive = &n
		}
	}
	if v := strings.TrimSpace(params.Get("search")); v != "" {
		f.Search = v
	}

	return f
}

// buildWhereConditions appends filter conditions to a WHERE clause using parameterized queries
// Returns the modified query and the collected args
func buildStudentWhere(base string, f ExportFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.Gender != "" {
		conditions = append(conditions, "gender = ?")
		args = append(args, f.Gender)
	}
	if f.EntryYear > 0 {
		conditions = append(conditions, "CAST(entry_year AS UNSIGNED) = ?")
		args = append(args, f.EntryYear)
	}
	if f.Search != "" {
		conditions = append(conditions, "(nis LIKE ? OR nisn LIKE ? OR full_name LIKE ? OR phone LIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like, like, like)
	}

	if len(conditions) == 0 {
		return base, args
	}

	return base + " AND " + strings.Join(conditions, " AND "), args
}

func buildTeacherWhere(base string, f ExportFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if f.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, f.Status)
	}
	if f.Gender != "" {
		conditions = append(conditions, "gender = ?")
		args = append(args, f.Gender)
	}
	if f.Search != "" {
		conditions = append(conditions, "(nip LIKE ? OR nuptk LIKE ? OR full_name LIKE ? OR email LIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like, like, like)
	}

	if len(conditions) == 0 {
		return base, args
	}

	return base + " AND " + strings.Join(conditions, " AND "), args
}

func buildDepartmentWhere(base string, f ExportFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if f.Search != "" {
		conditions = append(conditions, "(code LIKE ? OR name LIKE ? OR program_name LIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like, like)
	}

	if len(conditions) == 0 {
		return base, args
	}

	return base + " AND " + strings.Join(conditions, " AND "), args
}

func buildGradeLevelWhere(base string, f ExportFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if f.Search != "" {
		conditions = append(conditions, "(code LIKE ? OR name LIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like)
	}

	if len(conditions) == 0 {
		return base, args
	}

	return base + " AND " + strings.Join(conditions, " AND "), args
}

func buildAcademicYearWhere(base string, f ExportFilters) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if f.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *f.IsActive)
	}
	if f.Search != "" {
		conditions = append(conditions, "name LIKE ?")
		like := "%" + f.Search + "%"
		args = append(args, like)
	}

	if len(conditions) == 0 {
		return base, args
	}

	return base + " AND " + strings.Join(conditions, " AND "), args
}

type exportedStudent struct {
	ID        uint64
	NIS       string
	NISN      sql.NullString
	FullName  string
	Gender    string
	BirthPlace sql.NullString
	BirthDate sql.NullTime
	Address   sql.NullString
	Phone     sql.NullString
	EntryYear int
	Status    string
	CreatedAt time.Time
}

func exportStudents(db *sql.DB, f ExportFilters) ([]byte, error) {
	ctx := context.Background()

	baseQuery := `SELECT id, nis, nisn, full_name, gender, birth_place, birth_date, address, phone,
		       CAST(entry_year AS UNSIGNED), status, created_at
		FROM students
		WHERE deleted_at IS NULL`

	query, args := buildStudentWhere(baseQuery, f)
	query += " ORDER BY full_name ASC, id DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []exportedStudent
	for rows.Next() {
		var s exportedStudent
		if err := rows.Scan(&s.ID, &s.NIS, &s.NISN, &s.FullName, &s.Gender, &s.BirthPlace, &s.BirthDate, &s.Address, &s.Phone, &s.EntryYear, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	sheet := "Students"
	file.SetSheetName("Sheet1", sheet)

	headers := []string{
		"ID",
		"NIS",
		"NISN",
		"Nama Lengkap",
		"Jenis Kelamin",
		"Tempat Lahir",
		"Tanggal Lahir",
		"Alamat",
		"No. HP",
		"Tahun Masuk",
		"Status",
		"Tanggal Dibuat",
	}

	for col, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}

	style, err := file.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err == nil {
		_ = file.SetRowStyle(sheet, 1, 1, style)
	}

	for i, s := range students {
		row := i + 2
		setCellUint64(file, sheet, row, 1, s.ID)
		setCellString(file, sheet, row, 2, s.NIS)
		setCellNullableString(file, sheet, row, 3, s.NISN)
		setCellString(file, sheet, row, 4, s.FullName)
		setCellString(file, sheet, row, 5, s.Gender)
		setCellNullableString(file, sheet, row, 6, s.BirthPlace)
		setCellNullableTime(file, sheet, row, 7, s.BirthDate)
		setCellNullableString(file, sheet, row, 8, s.Address)
		setCellNullableString(file, sheet, row, 9, s.Phone)
		setCellInt(file, sheet, row, 10, s.EntryYear)
		setCellString(file, sheet, row, 11, s.Status)
		setCellTime(file, sheet, row, 12, s.CreatedAt)
	}

	_ = file.SetColWidth(sheet, "A", "A", 8)
	_ = file.SetColWidth(sheet, "B", "B", 15)
	_ = file.SetColWidth(sheet, "C", "C", 15)
	_ = file.SetColWidth(sheet, "D", "D", 25)
	_ = file.SetColWidth(sheet, "E", "E", 14)
	_ = file.SetColWidth(sheet, "F", "F", 18)
	_ = file.SetColWidth(sheet, "G", "G", 14)
	_ = file.SetColWidth(sheet, "H", "H", 30)
	_ = file.SetColWidth(sheet, "I", "I", 18)
	_ = file.SetColWidth(sheet, "J", "J", 12)
	_ = file.SetColWidth(sheet, "K", "K", 12)
	_ = file.SetColWidth(sheet, "L", "L", 20)

	_ = file.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomRight",
	})

	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func setCellString(f *excelize.File, sheet string, row, col int, val string) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, val)
}

func setCellUint64(f *excelize.File, sheet string, row, col int, val uint64) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, val)
}

func setCellInt(f *excelize.File, sheet string, row, col int, val int) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, val)
}

func setCellNullableString(f *excelize.File, sheet string, row, col int, val sql.NullString) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	if val.Valid {
		_ = f.SetCellValue(sheet, cell, val.String)
	}
}

func setCellNullableTime(f *excelize.File, sheet string, row, col int, val sql.NullTime) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	if val.Valid {
		_ = f.SetCellValue(sheet, cell, val.Time.Format("2006-01-02"))
	}
}

func setCellTime(f *excelize.File, sheet string, row, col int, val time.Time) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, val.Format("2006-01-02 15:04:05"))
}

func nullIfEmptyExport(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
