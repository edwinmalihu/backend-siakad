package importexport

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

var departmentsHeaders = []string{
	"Kode",
	"Nama",
	"Nama Program",
	"Bidang Keahlian",
	"Deskripsi",
}

var departmentsSampleRow = []string{
	"TKJ",
	"Teknik Komputer dan Jaringan",
	"Teknologi Informasi",
	"Jaringan Komputer",
	"Program keahlian bidang teknologi informasi dan komunikasi",
}

func generateDepartmentsTemplate() ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Template"
	f.SetSheetName("Sheet1", sheet)

	for col, header := range departmentsHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}

	for col, val := range departmentsSampleRow {
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

	_ = f.SetColWidth(sheet, "A", "A", 12)
	_ = f.SetColWidth(sheet, "B", "B", 35)
	_ = f.SetColWidth(sheet, "C", "C", 30)
	_ = f.SetColWidth(sheet, "D", "D", 28)
	_ = f.SetColWidth(sheet, "E", "E", 50)

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

func importDepartments(db *sql.DB, fileBytes []byte) (*ImportResult, error) {
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

	existingCode, existingName, err := loadExistingDepartmentIdentifiers(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing identifiers: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO departments (code, name, program_name, field_name, description)
		VALUES (?, ?, ?, ?, ?)
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

		if len(row) < 5 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("row has only %d columns, expected 5", len(row)),
			})
			result.ErrorCount++
			continue
		}

		code := strings.ToUpper(strings.TrimSpace(row[0]))
		name := strings.TrimSpace(row[1])
		programName := strings.TrimSpace(row[2])
		fieldName := strings.TrimSpace(row[3])
		description := strings.TrimSpace(row[4])

		var validationErrors []string

		if code == "" {
			validationErrors = append(validationErrors, "kode wajib diisi")
		}
		if name == "" {
			validationErrors = append(validationErrors, "nama wajib diisi")
		}

		if code != "" && existingCode[code] {
			validationErrors = append(validationErrors, fmt.Sprintf("kode '%s' sudah ada di database", code))
		}
		if name != "" && existingName[strings.ToLower(name)] {
			validationErrors = append(validationErrors, fmt.Sprintf("nama '%s' sudah ada di database", name))
		}

		if len(validationErrors) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: strings.Join(validationErrors, "; "),
			})
			result.ErrorCount++
			continue
		}

		_, err = stmt.ExecContext(ctx, code, name, nullIfEmpty(programName), nullIfEmpty(fieldName), nullIfEmpty(description))
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("database error: %s", err.Error()),
			})
			result.ErrorCount++
			continue
		}

		existingCode[code] = true
		existingName[strings.ToLower(name)] = true

		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func loadExistingDepartmentIdentifiers(db *sql.DB) (codeSet, nameSet map[string]bool, err error) {
	codeSet = make(map[string]bool)
	nameSet = make(map[string]bool)

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT code, name FROM departments WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			return nil, nil, err
		}
		codeSet[strings.ToUpper(strings.TrimSpace(code))] = true
		nameSet[strings.ToLower(strings.TrimSpace(name))] = true
	}

	return codeSet, nameSet, rows.Err()
}

func exportDepartments(db *sql.DB, f ExportFilters) ([]byte, error) {
	ctx := context.Background()

	baseQuery := `SELECT code, name, program_name, field_name, description
		FROM departments
		WHERE deleted_at IS NULL`

	query, args := buildDepartmentWhere(baseQuery, f)
	query += " ORDER BY code ASC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type exportedDepartment struct {
		Code        string
		Name        string
		ProgramName sql.NullString
		FieldName   sql.NullString
		Description sql.NullString
	}

	var departments []exportedDepartment
	for rows.Next() {
		var d exportedDepartment
		if err := rows.Scan(&d.Code, &d.Name, &d.ProgramName, &d.FieldName, &d.Description); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	sheet := "Departments"
	file.SetSheetName("Sheet1", sheet)

	headers := []string{"Kode", "Nama", "Nama Program", "Bidang Keahlian", "Deskripsi"}

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

	for i, d := range departments {
		row := i + 2
		setCellString(file, sheet, row, 1, d.Code)
		setCellString(file, sheet, row, 2, d.Name)
		setCellNullableString(file, sheet, row, 3, d.ProgramName)
		setCellNullableString(file, sheet, row, 4, d.FieldName)
		setCellNullableString(file, sheet, row, 5, d.Description)
	}

	_ = file.SetColWidth(sheet, "A", "A", 12)
	_ = file.SetColWidth(sheet, "B", "B", 35)
	_ = file.SetColWidth(sheet, "C", "C", 30)
	_ = file.SetColWidth(sheet, "D", "D", 28)
	_ = file.SetColWidth(sheet, "E", "E", 50)

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
