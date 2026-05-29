-- Add logout_time column to audit_logs for session tracking
-- Idempotent: only adds if column doesn't exist

SET @col_exists = (SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'audit_logs' AND COLUMN_NAME = 'logout_time');
SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `audit_logs` ADD COLUMN `logout_time` DATETIME DEFAULT NULL AFTER `ip_address`',
  'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'audit_logs' AND INDEX_NAME = 'idx_audit_logs_logout_time');
SET @sql = IF(@idx_exists = 0,
  'ALTER TABLE `audit_logs` ADD INDEX `idx_audit_logs_logout_time` (`logout_time`)',
  'SELECT 1');
PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;
