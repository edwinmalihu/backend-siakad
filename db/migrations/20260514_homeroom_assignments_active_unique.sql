-- Allow reusing homeroom assignment class scope after soft delete.
-- Also preserve the historical migration pattern by only enforcing class scope
-- uniqueness for active rows.

SET @db_name = DATABASE();

SELECT COUNT(*) INTO @ha_active_class_column_exists
FROM information_schema.columns
WHERE table_schema = @db_name
  AND table_name = 'homeroom_assignments'
  AND column_name = 'active_class_id';

SET @sql = IF(
  @ha_active_class_column_exists = 0,
  "ALTER TABLE `homeroom_assignments`
     ADD COLUMN `active_class_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
       CASE
         WHEN `deleted_at` IS NULL THEN `class_id`
         ELSE NULL
       END
     ) VIRTUAL AFTER `deleted_at`,
     ADD COLUMN `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
       CASE
         WHEN `deleted_at` IS NULL THEN `academic_year_id`
         ELSE NULL
       END
     ) VIRTUAL AFTER `active_class_id`,
     ADD COLUMN `active_semester_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
       CASE
         WHEN `deleted_at` IS NULL THEN `semester_id`
         ELSE NULL
       END
     ) VIRTUAL AFTER `active_academic_year_id`",
  "SELECT 1"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @ha_class_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'homeroom_assignments'
  AND index_name = 'idx_homeroom_assignments_class_id';

SET @sql = IF(
  @ha_class_index_exists = 0,
  "ALTER TABLE `homeroom_assignments` ADD KEY `idx_homeroom_assignments_class_id` (`class_id`)",
  "SELECT 1"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @ha_old_scope_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'homeroom_assignments'
  AND index_name = 'uk_homeroom_assignments_scope';

SET @sql = IF(@ha_old_scope_index_exists > 0, "ALTER TABLE `homeroom_assignments` DROP INDEX `uk_homeroom_assignments_scope`", "SELECT 1");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @ha_new_scope_index_exists
FROM information_schema.statistics
WHERE table_schema = @db_name
  AND table_name = 'homeroom_assignments'
  AND index_name = 'uk_homeroom_assignments_active_scope';

SET @sql = IF(
  @ha_new_scope_index_exists = 0,
  "ALTER TABLE `homeroom_assignments`
     ADD UNIQUE KEY `uk_homeroom_assignments_active_scope` (`active_class_id`, `active_academic_year_id`, `active_semester_id`)",
  "SELECT 1"
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
