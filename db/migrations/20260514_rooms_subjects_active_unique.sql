-- Make rooms and subjects reusable after soft delete.
-- This migration is written to be safe when run more than once on environments
-- that may already have some of these generated columns or indexes.

SET @db_name = DATABASE();

SELECT COUNT(*) INTO @room_active_code_column_exists
FROM information_schema.columns
WHERE table_schema = @db_name
  AND table_name = 'rooms'
  AND column_name = 'active_code';

SET @sql = IF(
  @room_active_code_column_exists = 0,
  "ALTER TABLE `rooms`
     ADD COLUMN `active_code` VARCHAR(30) GENERATED ALWAYS AS (
       CASE
         WHEN `deleted_at` IS NULL THEN `code`
         ELSE NULL
       END
     ) VIRTUAL AFTER `deleted_at`,
     ADD COLUMN `active_name` VARCHAR(100) GENERATED ALWAYS AS (
       CASE
         WHEN `deleted_at` IS NULL THEN `name`
         ELSE NULL
       END
     ) VIRTUAL AFTER `active_code`",
  "SELECT 1"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @room_old_code_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'rooms'
  AND index_name = 'uk_rooms_code';

SET @sql = IF(@room_old_code_index_exists > 0, "ALTER TABLE `rooms` DROP INDEX `uk_rooms_code`", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @room_old_name_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'rooms'
  AND index_name = 'uk_rooms_name';

SET @sql = IF(@room_old_name_index_exists > 0, "ALTER TABLE `rooms` DROP INDEX `uk_rooms_name`", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @room_new_code_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'rooms'
  AND index_name = 'uk_rooms_active_code';

SET @sql = IF(@room_new_code_index_exists = 0, "ALTER TABLE `rooms` ADD UNIQUE KEY `uk_rooms_active_code` (`active_code`)", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @room_new_name_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'rooms'
  AND index_name = 'uk_rooms_active_name';

SET @sql = IF(@room_new_name_index_exists = 0, "ALTER TABLE `rooms` ADD UNIQUE KEY `uk_rooms_active_name` (`active_name`)", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @subject_active_code_column_exists
FROM information_schema.columns
WHERE table_schema = @db_name
  AND table_name = 'subjects'
  AND column_name = 'active_code';

SET @sql = IF(
  @subject_active_code_column_exists = 0,
  "ALTER TABLE `subjects`
     ADD COLUMN `active_code` VARCHAR(30) GENERATED ALWAYS AS (
       CASE
         WHEN `deleted_at` IS NULL THEN `code`
         ELSE NULL
       END
     ) VIRTUAL AFTER `deleted_at`",
  "SELECT 1"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @subject_old_code_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'subjects'
  AND index_name = 'uk_subjects_code';

SET @sql = IF(@subject_old_code_index_exists > 0, "ALTER TABLE `subjects` DROP INDEX `uk_subjects_code`", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @subject_new_code_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'subjects'
  AND index_name = 'uk_subjects_active_code';

SET @sql = IF(@subject_new_code_index_exists = 0, "ALTER TABLE `subjects` ADD UNIQUE KEY `uk_subjects_active_code` (`active_code`)", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
