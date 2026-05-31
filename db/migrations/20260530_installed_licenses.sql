-- License Generator: installed_licenses table
-- Stores local license state on the SIAKAD client side.

CREATE TABLE IF NOT EXISTS installed_licenses (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  license_key VARCHAR(100) NOT NULL,
  tier ENUM('trial', 'enterprise') NOT NULL,
  device_fingerprint VARCHAR(255) NOT NULL,
  starts_at DATETIME NOT NULL,
  expires_at DATETIME NOT NULL,
  trial_count INT DEFAULT 0,
  client_name VARCHAR(200) DEFAULT NULL,
  last_alert_at DATETIME DEFAULT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
