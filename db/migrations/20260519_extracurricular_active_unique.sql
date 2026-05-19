ALTER TABLE `extracurriculars`
  DROP INDEX `uk_extracurriculars_name`,
  ADD COLUMN `active_name` VARCHAR(150)
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL THEN `name`
        ELSE NULL
      END
    ) VIRTUAL
    AFTER `deleted_at`,
  ADD UNIQUE KEY `uk_extracurriculars_active_name` (`active_name`);

ALTER TABLE `extracurricular_members`
  DROP INDEX `uk_extracurricular_members_scope`,
  ADD COLUMN `active_extracurricular_id` BIGINT UNSIGNED
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL THEN `extracurricular_id`
        ELSE NULL
      END
    ) VIRTUAL
    AFTER `deleted_at`,
  ADD COLUMN `active_student_id` BIGINT UNSIGNED
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL THEN `student_id`
        ELSE NULL
      END
    ) VIRTUAL
    AFTER `active_extracurricular_id`,
  ADD COLUMN `active_academic_year_id` BIGINT UNSIGNED
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL THEN `academic_year_id`
        ELSE NULL
      END
    ) VIRTUAL
    AFTER `active_student_id`,
  ADD UNIQUE KEY `uk_extracurricular_members_active_scope` (`active_extracurricular_id`, `active_student_id`, `active_academic_year_id`),
  ADD KEY `idx_extracurricular_members_extracurricular_id` (`extracurricular_id`);
