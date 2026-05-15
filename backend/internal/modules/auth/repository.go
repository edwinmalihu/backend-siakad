package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("user is inactive")
	ErrUserNotFound       = errors.New("user not found")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByIdentifier(ctx context.Context, identifier string) (*User, error) {
	const query = `
		SELECT
			u.id,
			u.username,
			COALESCE(u.email, ''),
			u.password_hash,
			u.is_active,
			u.last_login_at,
			COALESCE(up.full_name, ''),
			COALESCE(up.phone, ''),
			COALESCE(up.photo_url, '')
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id AND up.deleted_at IS NULL
		WHERE (u.username = ? OR u.email = ?)
		  AND u.deleted_at IS NULL
		LIMIT 1
	`

	identifier = normalizeIdentifier(identifier)

	var user User
	var email sql.NullString
	var lastLoginAt sql.NullTime
	var fullName sql.NullString
	var phone sql.NullString
	var photoURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, identifier, identifier).Scan(
		&user.ID,
		&user.Username,
		&email,
		&user.PasswordHash,
		&user.IsActive,
		&lastLoginAt,
		&fullName,
		&phone,
		&photoURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user by identifier: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if fullName.Valid {
		user.FullName = fullName.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	user.RoleCodes = make([]string, 0, len(roles))
	for _, role := range roles {
		user.RoleCodes = append(user.RoleCodes, role.Code)
	}

	return &user, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint64) (*User, error) {
	const query = `
		SELECT
			u.id,
			u.username,
			COALESCE(u.email, ''),
			u.password_hash,
			u.is_active,
			u.last_login_at,
			COALESCE(up.full_name, ''),
			COALESCE(up.phone, ''),
			COALESCE(up.photo_url, '')
		FROM users u
		LEFT JOIN user_profiles up ON up.user_id = u.id AND up.deleted_at IS NULL
		WHERE u.id = ?
		  AND u.deleted_at IS NULL
		LIMIT 1
	`

	var user User
	var email sql.NullString
	var lastLoginAt sql.NullTime
	var fullName sql.NullString
	var phone sql.NullString
	var photoURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&email,
		&user.PasswordHash,
		&user.IsActive,
		&lastLoginAt,
		&fullName,
		&phone,
		&photoURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	if fullName.Valid {
		user.FullName = fullName.String
	}
	if phone.Valid {
		user.Phone = phone.String
	}
	if photoURL.Valid {
		user.PhotoURL = photoURL.String
	}

	roles, err := r.loadRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Roles = roles
	user.RoleCodes = make([]string, 0, len(roles))
	for _, role := range roles {
		user.RoleCodes = append(user.RoleCodes, role.Code)
	}

	return &user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET last_login_at = NOW()
		WHERE id = ? AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}

func (r *Repository) loadRoles(ctx context.Context, userID uint64) ([]Role, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.name, r.code
		FROM user_roles ur
		INNER JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = ?
		  AND r.deleted_at IS NULL
		ORDER BY r.name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("load user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]Role, 0)
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Code); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user roles: %w", err)
	}

	return roles, nil
}
