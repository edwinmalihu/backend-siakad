package license

import (
	"context"
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetActive(ctx context.Context) (*InstalledLicense, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, license_key, tier, device_fingerprint, starts_at, expires_at,
		        trial_count, client_name, last_alert_at, created_at, updated_at
		 FROM installed_licenses
		 ORDER BY id DESC LIMIT 1`)

	var l InstalledLicense
	err := row.Scan(&l.ID, &l.LicenseKey, &l.Tier, &l.DeviceFingerprint,
		&l.StartsAt, &l.ExpiresAt, &l.TrialCount, &l.ClientName,
		&l.LastAlertAt, &l.CreatedAt, &l.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *Repository) Upsert(ctx context.Context, l *InstalledLicense) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO installed_licenses (license_key, tier, device_fingerprint, starts_at, expires_at, trial_count, client_name)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   license_key = VALUES(license_key),
		   tier = VALUES(tier),
		   starts_at = VALUES(starts_at),
		   expires_at = VALUES(expires_at),
		   trial_count = VALUES(trial_count),
		   client_name = VALUES(client_name),
		   updated_at = NOW()`,
		l.LicenseKey, l.Tier, l.DeviceFingerprint, l.StartsAt, l.ExpiresAt, l.TrialCount, l.ClientName)
	return err
}

func (r *Repository) Save(ctx context.Context, l *InstalledLicense) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO installed_licenses (license_key, tier, device_fingerprint, starts_at, expires_at, trial_count, client_name)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		l.LicenseKey, l.Tier, l.DeviceFingerprint, l.StartsAt, l.ExpiresAt, l.TrialCount, l.ClientName)
	return err
}

func (r *Repository) UpdateLastAlert(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE installed_licenses SET last_alert_at = NOW() WHERE id = ?`, id)
	return err
}

func (r *Repository) GetTrialCount(ctx context.Context, deviceFingerprint string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(trial_count), 0) FROM installed_licenses WHERE device_fingerprint = ?`,
		deviceFingerprint).Scan(&count)
	return count, err
}
