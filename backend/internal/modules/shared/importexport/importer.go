package importexport

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type ImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ImportResult struct {
	TotalRows    int            `json:"total_rows"`
	SuccessCount int            `json:"success_count"`
	ErrorCount   int            `json:"error_count"`
	Errors       []ImportError  `json:"errors"`
}

func importStudents(db *sql.DB, fileBytes []byte) (*ImportResult, error) {
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

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO students (nis, nisn, full_name, gender, birth_place, birth_date, address, phone, entry_year, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range rows {
		if i == 0 {
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

		nis := strings.ToUpper(strings.TrimSpace(row[0]))
		nisn := strings.TrimSpace(row[1])
		fullName := strings.TrimSpace(row[2])
		gender := strings.ToLower(strings.TrimSpace(row[3]))
		birthPlace := strings.TrimSpace(row[4])
		birthDateStr := strings.TrimSpace(row[5])
		address := strings.TrimSpace(row[6])
		phone := strings.TrimSpace(row[7])
		entryYearStr := strings.TrimSpace(row[8])
		status := strings.ToLower(strings.TrimSpace(row[9]))

		var validationErrors []string

		if nis == "" {
			validationErrors = append(validationErrors, "NIS is required")
		}
		if fullName == "" {
			validationErrors = append(validationErrors, "nama lengkap is required")
		}
		if gender != "male" && gender != "female" {
			validationErrors = append(validationErrors, "jenis kelamin must be 'male' or 'female'")
		}

		entryYear, err := strconv.Atoi(entryYearStr)
		if err != nil || entryYear < 1901 || entryYear > 2155 {
			validationErrors = append(validationErrors, "tahun masuk must be a valid year (1901-2155)")
		}

		if status == "" {
			status = "active"
		}
		if status != "active" && status != "inactive" {
			validationErrors = append(validationErrors, "status must be 'active' or 'inactive'")
		}

		var parsedBirthDate *time.Time
		if birthDateStr != "" {
			parsed, err := time.Parse("2006-01-02", birthDateStr)
			if err != nil {
				validationErrors = append(validationErrors, "tanggal lahir must use YYYY-MM-DD format")
			} else {
				parsedBirthDate = &parsed
			}
		}

		if len(validationErrors) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: strings.Join(validationErrors, "; "),
			})
			result.ErrorCount++
			continue
		}

		_, err = stmt.ExecContext(nil, nis, nullIfEmpty(nisn), fullName, gender, nullIfEmpty(birthPlace), parsedBirthDate, nullIfEmpty(address), nullIfEmpty(phone), entryYear, status)
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "uk_students_nis") {
				validationErrors = append(validationErrors, fmt.Sprintf("NIS '%s' sudah ada di database", nis))
			} else if strings.Contains(errMsg, "uk_students_nisn") {
				validationErrors = append(validationErrors, fmt.Sprintf("NISN '%s' sudah ada di database", nisn))
			} else {
				validationErrors = append(validationErrors, fmt.Sprintf("database error: %s", errMsg))
			}
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: strings.Join(validationErrors, "; "),
			})
			result.ErrorCount++
			continue
		}

		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
