-- Token revocation table for logout
CREATE TABLE IF NOT EXISTS revoked_tokens (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  token_hash VARCHAR(64) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  revoked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_revoked_tokens_hash (token_hash),
  KEY idx_revoked_tokens_user (user_id),
  KEY idx_revoked_tokens_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Cleanup job: delete expired revoked tokens periodically
-- (can be run via cron or scheduled event)
-- DELETE FROM revoked_tokens WHERE expires_at < NOW();
