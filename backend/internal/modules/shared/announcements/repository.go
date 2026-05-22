package announcements

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("announcement not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, search, targetScope string, isPublished *bool) ([]Announcement, error) {
	query := `
		SELECT
			id,
			created_by,
			title,
			content,
			target_scope,
			publish_start,
			publish_end,
			is_published,
			created_at,
			updated_at,
			deleted_at
		FROM announcements
		WHERE deleted_at IS NULL
	`
	args := make([]any, 0, 4)
	if search != "" {
		query += " AND (title LIKE ? OR content LIKE ?)"
		pattern := "%" + strings.TrimSpace(search) + "%"
		args = append(args, pattern, pattern)
	}
	if targetScope != "" {
		query += " AND target_scope = ?"
		args = append(args, strings.TrimSpace(targetScope))
	}
	if isPublished != nil {
		query += " AND is_published = ?"
		args = append(args, *isPublished)
	}
	query += " ORDER BY is_published DESC, COALESCE(publish_start, created_at) DESC, id DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query announcements: %w", err)
	}
	defer rows.Close()

	items := make([]Announcement, 0)
	for rows.Next() {
		item, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate announcements: %w", err)
	}
	return items, nil
}

func (r *Repository) GetByID(ctx context.Context, id uint64) (*Announcement, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			created_by,
			title,
			content,
			target_scope,
			publish_start,
			publish_end,
			is_published,
			created_at,
			updated_at,
			deleted_at
		FROM announcements
		WHERE id = ? AND deleted_at IS NULL
		LIMIT 1
	`, id)
	item, err := scanAnnouncement(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get announcement by id: %w", err)
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, item Announcement) (*Announcement, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO announcements (created_by, title, content, target_scope, publish_start, publish_end, is_published)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, nullableUint64(item.CreatedBy), item.Title, item.Content, nullableString(item.TargetScope), nullableTime(item.PublishStart), nullableTime(item.PublishEnd), item.IsPublished)
	if err != nil {
		return nil, fmt.Errorf("insert announcement: %w", err)
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted announcement id: %w", err)
	}
	return r.GetByID(ctx, uint64(insertedID))
}

func (r *Repository) Update(ctx context.Context, id uint64, item Announcement) (*Announcement, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE announcements
		SET title = ?, content = ?, target_scope = ?, publish_start = ?, publish_end = ?, is_published = ?
		WHERE id = ? AND deleted_at IS NULL
	`, item.Title, item.Content, nullableString(item.TargetScope), nullableTime(item.PublishStart), nullableTime(item.PublishEnd), item.IsPublished, id)
	if err != nil {
		return nil, fmt.Errorf("update announcement: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("get updated announcement affected rows: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE announcements
		SET deleted_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("soft delete announcement: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted announcement affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAnnouncement(s scanner) (Announcement, error) {
	var item Announcement
	var createdBy sql.NullInt64
	var targetScope sql.NullString
	var publishStart sql.NullTime
	var publishEnd sql.NullTime
	err := s.Scan(&item.ID, &createdBy, &item.Title, &item.Content, &targetScope, &publishStart, &publishEnd, &item.IsPublished, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if err != nil {
		return Announcement{}, err
	}
	if createdBy.Valid {
		value := uint64(createdBy.Int64)
		item.CreatedBy = &value
	}
	if targetScope.Valid {
		item.TargetScope = targetScope.String
	}
	if publishStart.Valid {
		value := publishStart.Time
		item.PublishStart = &value
	}
	if publishEnd.Valid {
		value := publishEnd.Time
		item.PublishEnd = &value
	}
	return item, nil
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02 15:04:05")
}

func nullableUint64(value *uint64) any {
	if value == nil || *value == 0 {
		return nil
	}
	return *value
}
