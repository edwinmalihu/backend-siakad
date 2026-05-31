package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"
)

type RevokedTokenRepository struct {
	db *sql.DB
}

func NewRevokedTokenRepository(db *sql.DB) *RevokedTokenRepository {
	return &RevokedTokenRepository{db: db}
}

// Revoke stores a token hash in the blacklist
func (r *RevokedTokenRepository) Revoke(ctx context.Context, token string, userID uint64, expiresAt time.Time) error {
	hash := hashToken(token)
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO revoked_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hash, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

// IsRevoked checks if a token has been revoked
func (r *RevokedTokenRepository) IsRevoked(ctx context.Context, token string) (bool, error) {
	hash := hashToken(token)
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE token_hash = ? AND expires_at > NOW())`,
		hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check revoked token: %w", err)
	}
	return exists, nil
}

// Cleanup removes expired revoked tokens
func (r *RevokedTokenRepository) Cleanup(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM revoked_tokens WHERE expires_at < NOW()`)
	if err != nil {
		return 0, fmt.Errorf("cleanup revoked tokens: %w", err)
	}
	return result.RowsAffected()
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
