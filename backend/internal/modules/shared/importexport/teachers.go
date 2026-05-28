package importexport

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

var teachersHeaders = []string{
	"NIP",
	"NUPTK",
	"Nama Lengkap",
	"Jenis Kelamin (male/female)",
	"Alamat",
	"No. HP",
	"Email",
	"Status Kepegawaian",
	"Jabatan",
	"Status (active/inactive)",
}

var teachersSampleRow = []string{
	"198501012010011001",
	"2015123456700001",
	"Dr. Ahmad Fauzi, M.Pd",
	"male",
	"Jl. Pendidikan No. 10",
	"081234567890",
	"ahmad.fauzi@sekolah.sch.id",
	"PNS",
	"Guru Matematika",
	"active",
}

func generateTeachersTemplate() ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Template"
	f.SetSheetName("Sheet1", sheet)

	for col, header := range teachersHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}

	for col, val := range teachersSampleRow {
		cell, _ := excelize.CoordinatesToCellName(col+1, 2)
		_ = f.SetCellValue(sheet, cell, val)
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

	_ = f.SetColWidth(sheet, "A", "A", 22)
	_ = f.SetColWidth(sheet, "B", "B", 22)
	_ = f.SetColWidth(sheet, "C", "C", 28)
	_ = f.SetColWidth(sheet, "D", "D", 24)
	_ = f.SetColWidth(sheet, "E", "E", 30)
	_ = f.SetColWidth(sheet, "F", "F", 18)
	_ = f.SetColWidth(sheet, "G", "G", 30)
	_ = f.SetColWidth(sheet, "H", "H", 20)
	_ = f.SetColWidth(sheet, "I", "I", 22)
	_ = f.SetColWidth(sheet, "J", "J", 18)

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomRight",
	})

	genderDV := excelize.NewDataValidation(true)
	genderDV.SetSqref("D2:D1048576")
	_ = genderDV.SetDropList([]string{"male", "female"})
	_ = f.AddDataValidation(sheet, genderDV)

	statusDV := excelize.NewDataValidation(true)
	statusDV.SetSqref("J2:J1048576")
	_ = statusDV.SetDropList([]string{"active", "inactive"})
	_ = f.AddDataValidation(sheet, statusDV)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func importTeachers(db *sql.DB, fileBytes []byte) (*ImportResult, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(fileBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("no sheet found in excel file")
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("template must have at least 1 data row (excluding header)")
	}

	result := &ImportResult{
		TotalRows: len(rows) - 1,
		Errors:    []ImportError{},
	}

	existingNIP, existingNUPTK, existingEmail, err := loadExistingTeacherIdentifiers(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing identifiers: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO teachers (nip, nuptk, full_name, gender, address, phone, email, employment_status, position, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	ctx := context.Background()

	for i, row := range rows {
		if i == 0 {
			continue
		}

		if isRowEmpty(row) {
			result.SkippedRows++
			continue
		}

		if len(row) < 10 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("row has only %d columns, expected 10", len(row)),
			})
			result.ErrorCount++
			continue
		}

		nip := strings.TrimSpace(row[0])
		nuptk := strings.TrimSpace(row[1])
		fullName := strings.TrimSpace(row[2])
		gender := strings.ToLower(strings.TrimSpace(row[3]))
		address := strings.TrimSpace(row[4])
		phone := strings.TrimSpace(row[5])
		email := strings.TrimSpace(row[6])
		employmentStatus := strings.TrimSpace(row[7])
		position := strings.TrimSpace(row[8])
		status := strings.ToLower(strings.TrimSpace(row[9]))

		var validationErrors []string

		if fullName == "" {
			validationErrors = append(validationErrors, "nama lengkap wajib diisi")
		}
		if gender != "" && gender != "male" && gender != "female" {
			validationErrors = append(validationErrors, "jenis kelamin harus 'male' atau 'female'")
		}
		if status == "" {
			status = "active"
		}
		if status != "active" && status != "inactive" {
			validationErrors = append(validationErrors, "status harus 'active' atau 'inactive'")
		}
		if email != "" && !strings.Contains(email, "@") {
			validationErrors = append(validationErrors, "format email tidak valid")
		}

		if nip != "" && existingNIP[nip] {
			validationErrors = append(validationErrors, fmt.Sprintf("NIP '%s' sudah ada di database", nip))
		}
		if nuptk != "" && existingNUPTK[nuptk] {
			validationErrors = append(validationErrors, fmt.Sprintf("NUPTK '%s' sudah ada di database", nuptk))
		}
		if email != "" && existingEmail[strings.ToLower(email)] {
			validationErrors = append(validationErrors, fmt.Sprintf("email '%s' sudah ada di database", email))
		}

		if len(validationErrors) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: strings.Join(validationErrors, "; "),
			})
			result.ErrorCount++
			continue
		}

		_, err = stmt.ExecContext(ctx,
			nullIfEmpty(nip), nullIfEmpty(nuptk), fullName,
			nullIfEmpty(gender), nullIfEmpty(address), nullIfEmpty(phone),
			nullIfEmpty(strings.ToLower(email)), nullIfEmpty(employmentStatus),
			nullIfEmpty(position), status,
		)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("database error: %s", err.Error()),
			})
			result.ErrorCount++
			continue
		}

		if nip != "" {
			existingNIP[nip] = true
		}
		if nuptk != "" {
			existingNUPTK[nuptk] = true
		}
		if email != "" {
			existingEmail[strings.ToLower(email)] = true
		}

		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func loadExistingTeacherIdentifiers(db *sql.DB) (nipSet, nuptkSet, emailSet map[string]bool, err error) {
	nipSet = make(map[string]bool)
	nuptkSet = make(map[string]bool)
	emailSet = make(map[string]bool)

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT nip, nuptk, email FROM teachers WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nip, nuptk, email sql.NullString
		if err := rows.Scan(&nip, &nuptk, &email); err != nil {
			return nil, nil, nil, err
		}
		if nip.Valid && strings.TrimSpace(nip.String) != "" {
			nipSet[strings.TrimSpace(nip.String)] = true
		}
		if nuptk.Valid && strings.TrimSpace(nuptk.String) != "" {
			nuptkSet[strings.TrimSpace(nuptk.String)] = true
		}
		if email.Valid && strings.TrimSpace(email.String) != "" {
			emailSet[strings.ToLower(strings.TrimSpace(email.String))] = true
		}
	}

	return nipSet, nuptkSet, emailSet, rows.Err()
}

func exportTeachers(db *sql.DB, f ExportFilters) ([]byte, error) {
	ctx := context.Background()

	baseQuery := `SELECT nip, nuptk, full_name, gender, address, phone, email, employment_status, position, status
		FROM teachers
		WHERE deleted_at IS NULL`

	query, args := buildTeacherWhere(baseQuery, f)
	query += " ORDER BY full_name ASC, id DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type exportedTeacher struct {
		NIP              sql.NullString
		NUPTK            sql.NullString
		FullName         string
		Gender           sql.NullString
		Address          sql.NullString
		Phone            sql.NullString
		Email            sql.NullString
		EmploymentStatus sql.NullString
		Position         sql.NullString
		Status           string
	}

	var teachers []exportedTeacher
	for rows.Next() {
		var t exportedTeacher
		if err := rows.Scan(&t.NIP, &t.NUPTK, &t.FullName, &t.Gender, &t.Address, &t.Phone, &t.Email, &t.EmploymentStatus, &t.Position, &t.Status); err != nil {
			return nil, err
		}
		teachers = append(teachers, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	sheet := "Teachers"
	file.SetSheetName("Sheet1", sheet)

	headers := []string{
		"NIP", "NUPTK", "Nama Lengkap", "Jenis Kelamin", "Alamat",
		"No. HP", "Email", "Status Kepegawaian", "Jabatan", "Status",
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

	for i, t := range teachers {
		row := i + 2
		setCellNullableString(file, sheet, row, 1, t.NIP)
		setCellNullableString(file, sheet, row, 2, t.NUPTK)
		setCellString(file, sheet, row, 3, t.FullName)
		setCellNullableString(file, sheet, row, 4, t.Gender)
		setCellNullableString(file, sheet, row, 5, t.Address)
		setCellNullableString(file, sheet, row, 6, t.Phone)
		setCellNullableString(file, sheet, row, 7, t.Email)
		setCellNullableString(file, sheet, row, 8, t.EmploymentStatus)
		setCellNullableString(file, sheet, row, 9, t.Position)
		setCellString(file, sheet, row, 10, t.Status)
	}

	_ = file.SetColWidth(sheet, "A", "A", 22)
	_ = file.SetColWidth(sheet, "B", "B", 22)
	_ = file.SetColWidth(sheet, "C", "C", 28)
	_ = file.SetColWidth(sheet, "D", "D", 16)
	_ = file.SetColWidth(sheet, "E", "E", 30)
	_ = file.SetColWidth(sheet, "F", "F", 18)
	_ = file.SetColWidth(sheet, "G", "G", 30)
	_ = file.SetColWidth(sheet, "H", "H", 20)
	_ = file.SetColWidth(sheet, "I", "I", 22)
	_ = file.SetColWidth(sheet, "J", "J", 14)

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
