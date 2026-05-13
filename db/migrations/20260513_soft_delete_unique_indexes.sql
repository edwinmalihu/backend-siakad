-- Allow reusing master unique values after soft delete by moving uniqueness
-- checks to generated columns that are only populated when deleted_at IS NULL.

ALTER TABLE `academic_years`
  ADD COLUMN `active_name` VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  DROP INDEX `uk_academic_years_name`,
  ADD UNIQUE KEY `uk_academic_years_active_name` (`active_name`);

ALTER TABLE `semesters`
  ADD COLUMN `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD COLUMN `active_code` VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_academic_year_id`,
  ADD KEY `idx_semesters_academic_year_id` (`academic_year_id`),
  DROP INDEX `uk_semesters_year_code`,
  ADD UNIQUE KEY `uk_semesters_active_year_code` (`active_academic_year_id`, `active_code`);

ALTER TABLE `departments`
  ADD COLUMN `active_code` VARCHAR(30) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD COLUMN `active_name` VARCHAR(150) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_code`,
  DROP INDEX `uk_departments_code`,
  DROP INDEX `uk_departments_name`,
  ADD UNIQUE KEY `uk_departments_active_code` (`active_code`),
  ADD UNIQUE KEY `uk_departments_active_name` (`active_name`);

ALTER TABLE `grade_levels`
  ADD COLUMN `active_code` VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD COLUMN `active_name` VARCHAR(50) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_code`,
  DROP INDEX `uk_grade_levels_code`,
  DROP INDEX `uk_grade_levels_name`,
  ADD UNIQUE KEY `uk_grade_levels_active_code` (`active_code`),
  ADD UNIQUE KEY `uk_grade_levels_active_name` (`active_name`);

ALTER TABLE `classes`
  ADD COLUMN `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD COLUMN `active_department_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `department_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_academic_year_id`,
  ADD COLUMN `active_grade_level_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `grade_level_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_department_id`,
  ADD COLUMN `active_name` VARCHAR(50) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_grade_level_id`,
  ADD KEY `idx_classes_academic_year_id` (`academic_year_id`),
  DROP INDEX `uk_classes_unique_scope`,
  ADD UNIQUE KEY `uk_classes_active_scope` (`active_academic_year_id`, `active_department_id`, `active_grade_level_id`, `active_name`);
