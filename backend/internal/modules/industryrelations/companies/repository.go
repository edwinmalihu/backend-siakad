package companies

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrNotFound         = errors.New("company not found")
	ErrCategoryNotFound = errors.New("industry category not found")
	ErrDuplicateName    = errors.New("company name already exists")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, city, status string, categoryID uint64) ([]Company, error) {
	query := `
		SELECT
			c.id,
			c.category_id,
			COALESCE(ic.name, ''),
			c.name,
			c.city,
			c.address,
			c.contact_person,
			c.phone,
			c.email,
			c.status,
			c.created_at,
			c.updated_at,
			c.deleted_at
		FROM companies c
		LEFT JOIN industry_categories ic ON ic.id = c.category_id AND ic.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
	`
	args := make([]any, 0, 8)
	if search != "" {
		query += " AND (c.name LIKE ? OR c.contact_person LIKE ? OR c.email LIKE ? OR ic.name LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if city != "" {
		query += " AND c.city = ?"
		args = append(args, strings.TrimSpace(city))
	}
	if status != "" {
		query += " AND c.status = ?"
		args = append(args, status)
	}
	if categoryID > 0 {
		query += " AND c.category_id = ?"
		args = append(args, categoryID)
	}
	query += " ORDER BY c.name ASC, c.id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query companies: %w", err)
	}
	defer rows.Close()

	items := make([]Company, 0)
	for rows.Next() {
		item, err := scanCompany(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate companies: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Company, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			c.id,
			c.category_id,
			COALESCE(ic.name, ''),
			c.name,
			c.city,
			c.address,
			c.contact_person,
			c.phone,
			c.email,
			c.status,
			c.created_at,
			c.updated_at,
			c.deleted_at
		FROM companies c
		LEFT JOIN industry_categories ic ON ic.id = c.category_id AND ic.deleted_at IS NULL
		WHERE c.id = ? AND c.deleted_at IS NULL
		LIMIT 1
	`, id)
	item, err := scanCompany(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get company by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Company) (*Company, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create company transaction: %w", err)
	}
	defer tx.Rollback()

	if err := ensureCategoryExists(ctx, tx, item.CategoryID); err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO companies (category_id, name, city, address, contact_person, phone, email, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, nullableUint64(item.CategoryID), item.Name, nullableString(item.City), nullableString(item.Address), nullableString(item.ContactPerson), nullableString(item.Phone), nullableString(item.Email), item.Status)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("insert company: %w", err)
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted company id: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create company transaction: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Company) (*Company, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update company transaction: %w", err)
	}
	defer tx.Rollback()
	if err := ensureCompanyExists(ctx, tx, id); err != nil {
		return nil, err
	}
	if err := ensureCategoryExists(ctx, tx, item.CategoryID); err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE companies
		SET category_id = ?, name = ?, city = ?, address = ?, contact_person = ?, phone = ?, email = ?, status = ?
		WHERE id = ? AND deleted_at IS NULL
	`, nullableUint64(item.CategoryID), item.Name, nullableString(item.City), nullableString(item.Address), nullableString(item.ContactPerson), nullableString(item.Phone), nullableString(item.Email), item.Status, id)
	if err != nil {
		if dupErr := mapDuplicateError(err); dupErr != nil {
			return nil, dupErr
		}
		return nil, fmt.Errorf("update company: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated company affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update company transaction: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE companies
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete company: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted company affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCompany(s scanner) (Company, error) {
	var item Company
	var categoryID sql.NullInt64
	var city sql.NullString
	var address sql.NullString
	var contactPerson sql.NullString
	var phone sql.NullString
	var email sql.NullString
	err := s.Scan(&item.ID, &categoryID, &item.CategoryName, &item.Name, &city, &address, &contactPerson, &phone, &email, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		return Company{}, err
	}
	if categoryID.Valid {
		value := uint64(categoryID.Int64)
		item.CategoryID = &value
	}
	if city.Valid {
		item.City = city.String
	}
	if address.Valid {
		item.Address = address.String
	}
	if contactPerson.Valid {
		item.ContactPerson = contactPerson.String
	}
	if phone.Valid {
		item.Phone = phone.String
	}
	if email.Valid {
		item.Email = email.String
	}
	return item, nil
}

func ensureCategoryExists(ctx context.Context, tx *sql.Tx, id *uint64) error {
	if id == nil || *id == 0 {
		return nil
	}
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM industry_categories WHERE id = ? AND deleted_at IS NULL)`, *id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check industry category existence: %w", err)
	}
	if !exists {
		return ErrCategoryNotFound
	}
	return nil
}

func ensureCompanyExists(ctx context.Context, tx *sql.Tx, id uint64) error {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM companies WHERE id = ? AND deleted_at IS NULL)`, id).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check company existence: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableUint64(value *uint64) any {
	if value == nil || *value == 0 {
		return nil
	}
	return *value
}

func mapDuplicateError(err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
		return nil
	}
	if strings.Contains(strings.ToLower(mysqlErr.Message), "uk_companies_active_name") {
		return ErrDuplicateName
	}
	return fmt.Errorf("duplicate company data: %w", err)
}
