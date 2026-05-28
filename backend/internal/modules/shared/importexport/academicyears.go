package importexport

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var academicYearsHeaders = []string{
	"Nama",
	"Tanggal Mulai (YYYY-MM-DD)",
	"Tanggal Akhir (YYYY-MM-DD)",
	"Aktif (0/1)",
}

var academicYearsSampleRow = []string{
	"2025/2026",
	"2025-07-01",
	"2026-06-30",
	"1",
}

func generateAcademicYearsTemplate() ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Template"
	f.SetSheetName("Sheet1", sheet)

	for col, header := range academicYearsHeaders {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}

	for col, val := range academicYearsSampleRow {
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

	_ = f.SetColWidth(sheet, "A", "A", 18)
	_ = f.SetColWidth(sheet, "B", "B", 28)
	_ = f.SetColWidth(sheet, "C", "C", 28)
	_ = f.SetColWidth(sheet, "D", "D", 12)

	_ = f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomRight",
	})

	activeDV := excelize.NewDataValidation(true)
	activeDV.SetSqref("D2:D1048576")
	_ = activeDV.SetDropList([]string{"0", "1"})
	_ = f.AddDataValidation(sheet, activeDV)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func importAcademicYears(db *sql.DB, fileBytes []byte) (*ImportResult, error) {
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

	existingName, err := loadExistingAcademicYearNames(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing identifiers: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO academic_years (name, start_date, end_date, is_active)
		VALUES (?, ?, ?, ?)
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

		if len(row) < 4 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("row has only %d columns, expected 4", len(row)),
			})
			result.ErrorCount++
			continue
		}

		name := strings.TrimSpace(row[0])
		startDateStr := strings.TrimSpace(row[1])
		endDateStr := strings.TrimSpace(row[2])
		activeStr := strings.TrimSpace(row[3])

		var validationErrors []string

		if name == "" {
			validationErrors = append(validationErrors, "nama wajib diisi")
		}

		if startDateStr == "" {
			validationErrors = append(validationErrors, "tanggal mulai wajib diisi")
		}
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil && startDateStr != "" {
			validationErrors = append(validationErrors, "tanggal mulai harus format YYYY-MM-DD")
		}

		endDateStr2 := endDateStr
		var endDate time.Time
		if endDateStr2 == "" {
			validationErrors = append(validationErrors, "tanggal akhir wajib diisi")
		} else {
			endDate, err = time.Parse("2006-01-02", endDateStr2)
			if err != nil {
				validationErrors = append(validationErrors, "tanggal akhir harus format YYYY-MM-DD")
			}
		}

		isActive := 0
		if activeStr != "" {
			parsed, err := strconv.Atoi(activeStr)
			if err != nil || (parsed != 0 && parsed != 1) {
				validationErrors = append(validationErrors, "aktif harus 0 atau 1")
			} else {
				isActive = parsed
			}
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

		_, err = stmt.ExecContext(ctx, name, startDate, endDate, isActive)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("database error: %s", err.Error()),
			})
			result.ErrorCount++
			continue
		}

		existingName[strings.ToLower(name)] = true

		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func loadExistingAcademicYearNames(db *sql.DB) (nameSet map[string]bool, err error) {
	nameSet = make(map[string]bool)

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT name FROM academic_years WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		nameSet[strings.ToLower(strings.TrimSpace(name))] = true
	}

	return nameSet, rows.Err()
}

func exportAcademicYears(db *sql.DB, f ExportFilters) ([]byte, error) {
	ctx := context.Background()

	baseQuery := `SELECT name, start_date, end_date, is_active
		FROM academic_years
		WHERE deleted_at IS NULL`

	query, args := buildAcademicYearWhere(baseQuery, f)
	query += " ORDER BY start_date DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type exportedAcademicYear struct {
		Name      string
		StartDate time.Time
		EndDate   time.Time
		IsActive  int
	}

	var years []exportedAcademicYear
	for rows.Next() {
		var y exportedAcademicYear
		if err := rows.Scan(&y.Name, &y.StartDate, &y.EndDate, &y.IsActive); err != nil {
			return nil, err
		}
		years = append(years, y)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	file := excelize.NewFile()
	sheet := "Academic Years"
	file.SetSheetName("Sheet1", sheet)

	headers := []string{"Nama", "Tanggal Mulai", "Tanggal Akhir", "Aktif"}

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

	for i, y := range years {
		row := i + 2
		setCellString(file, sheet, row, 1, y.Name)
		setCellTime(file, sheet, row, 2, y.StartDate)
		setCellTime(file, sheet, row, 3, y.EndDate)
		setCellInt(file, sheet, row, 4, y.IsActive)
	}

	_ = file.SetColWidth(sheet, "A", "A", 18)
	_ = file.SetColWidth(sheet, "B", "B", 16)
	_ = file.SetColWidth(sheet, "C", "C", 16)
	_ = file.SetColWidth(sheet, "D", "D", 10)

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
