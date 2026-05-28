-- Add logout_time column to audit_logs for session tracking
-- On login: record is created with created_at as login_time
-- On logout: logout_time is updated on the same record

ALTER TABLE `audit_logs`
  ADD COLUMN `logout_time` DATETIME DEFAULT NULL AFTER `ip_address`;

-- Add index for faster session lookups
ALTER TABLE `audit_logs`
  ADD INDEX `idx_audit_logs_logout_time` (`logout_time`);
