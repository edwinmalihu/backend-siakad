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

type ImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ImportResult struct {
	TotalRows    int           `json:"total_rows"`
	SkippedRows  int           `json:"skipped_rows"`
	SuccessCount int           `json:"success_count"`
	ErrorCount   int           `json:"error_count"`
	Errors       []ImportError `json:"errors"`
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

	// pre-load existing NIS and NISN to check duplicates before insert
	existingNIS, existingNISN, err := loadExistingIdentifiers(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing identifiers: %w", err)
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

	ctx := context.Background()

	for i, row := range rows {
		if i == 0 {
			continue
		}

		// skip baris kosong — semua cell kosong atau whitespace
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

		// required field validation
		if nis == "" {
			validationErrors = append(validationErrors, "NIS wajib diisi")
		}
		if fullName == "" {
			validationErrors = append(validationErrors, "nama lengkap wajib diisi")
		}
		if gender != "male" && gender != "female" {
			validationErrors = append(validationErrors, "jenis kelamin harus 'male' atau 'female'")
		}

		entryYear, err := strconv.Atoi(entryYearStr)
		if err != nil || entryYear < 1901 || entryYear > 2155 {
			validationErrors = append(validationErrors, "tahun masuk harus tahun yang valid (1901-2155)")
		}

		if status == "" {
			status = "active"
		}
		if status != "active" && status != "inactive" {
			validationErrors = append(validationErrors, "status harus 'active' atau 'inactive'")
		}

		var parsedBirthDate *time.Time
		if birthDateStr != "" {
			parsed, err := time.Parse("2006-01-02", birthDateStr)
			if err != nil {
				validationErrors = append(validationErrors, "tanggal lahir harus format YYYY-MM-DD")
			} else {
				parsedBirthDate = &parsed
			}
		}

		// pre-check duplikat NIS
		if nis != "" && existingNIS[nis] {
			validationErrors = append(validationErrors, fmt.Sprintf("NIS '%s' sudah ada di database", nis))
		}

		// pre-check duplikat NISN
		if nisn != "" && existingNISN[nisn] {
			validationErrors = append(validationErrors, fmt.Sprintf("NISN '%s' sudah ada di database", nisn))
		}

		if len(validationErrors) > 0 {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: strings.Join(validationErrors, "; "),
			})
			result.ErrorCount++
			continue
		}

		_, err = stmt.ExecContext(ctx, nis, nullIfEmpty(nisn), fullName, gender, nullIfEmpty(birthPlace), parsedBirthDate, nullIfEmpty(address), nullIfEmpty(phone), entryYear, status)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				Message: fmt.Sprintf("database error: %s", err.Error()),
			})
			result.ErrorCount++
			continue
		}

		// add to existing set supaya cross-check dalam 1 file juga terdeteksi
		existingNIS[nis] = true
		if nisn != "" {
			existingNISN[nisn] = true
		}

		result.SuccessCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// loadExistingIdentifiers mengambil semua NIS dan NISN yang sudah ada di DB
func loadExistingIdentifiers(db *sql.DB) (nisSet map[string]bool, nisnSet map[string]bool, err error) {
	nisSet = make(map[string]bool)
	nisnSet = make(map[string]bool)

	ctx := context.Background()
	rows, err := db.QueryContext(ctx, `SELECT nis, nisn FROM students WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var nis string
		var nisn sql.NullString
		if err := rows.Scan(&nis, &nisn); err != nil {
			return nil, nil, err
		}
		nisSet[strings.ToUpper(nis)] = true
		if nisn.Valid && strings.TrimSpace(nisn.String) != "" {
			nisnSet[strings.TrimSpace(nisn.String)] = true
		}
	}

	return nisSet, nisnSet, rows.Err()
}

// isRowEmpty mengecek apakah semua cell dalam row kosong atau whitespace
func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func nullIfEmpty(s string) interface{} {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
