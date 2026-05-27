package importexport

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

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

func exportStudents(db *sql.DB) ([]byte, error) {
	ctx := context.Background()

	rows, err := db.QueryContext(ctx, `
		SELECT id, nis, nisn, full_name, gender, birth_place, birth_date, address, phone,
		       CAST(entry_year AS UNSIGNED), status, created_at
		FROM students
		WHERE deleted_at IS NULL
		ORDER BY full_name ASC, id DESC
	`)
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

	f := excelize.NewFile()
	sheet := "Students"
	f.SetSheetName("Sheet1", sheet)

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
		_ = f.SetCellValue(sheet, cell, header)
	}

	style, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err == nil {
		_ = f.SetRowStyle(sheet, 1, 1, style)
	}

	for i, s := range students {
		row := i + 2
		setCellUint64(f, sheet, row, 1, s.ID)
		setCellString(f, sheet, row, 2, s.NIS)
		setCellNullableString(f, sheet, row, 3, s.NISN)
		setCellString(f, sheet, row, 4, s.FullName)
		setCellString(f, sheet, row, 5, s.Gender)
		setCellNullableString(f, sheet, row, 6, s.BirthPlace)
		setCellNullableTime(f, sheet, row, 7, s.BirthDate)
		setCellNullableString(f, sheet, row, 8, s.Address)
		setCellNullableString(f, sheet, row, 9, s.Phone)
		setCellInt(f, sheet, row, 10, s.EntryYear)
		setCellString(f, sheet, row, 11, s.Status)
		setCellTime(f, sheet, row, 12, s.CreatedAt)
	}

	_ = f.SetColWidth(sheet, "A", "A", 8)
	_ = f.SetColWidth(sheet, "B", "B", 15)
	_ = f.SetColWidth(sheet, "C", "C", 15)
	_ = f.SetColWidth(sheet, "D", "D", 25)
	_ = f.SetColWidth(sheet, "E", "E", 14)
	_ = f.SetColWidth(sheet, "F", "F", 18)
	_ = f.SetColWidth(sheet, "G", "G", 14)
	_ = f.SetColWidth(sheet, "H", "H", 30)
	_ = f.SetColWidth(sheet, "I", "I", 18)
	_ = f.SetColWidth(sheet, "J", "J", 12)
	_ = f.SetColWidth(sheet, "K", "K", 12)
	_ = f.SetColWidth(sheet, "L", "L", 20)

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomRight",
	})

	buf, err := f.WriteToBuffer()
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
