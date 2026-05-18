ALTER TABLE `student_enrollments`
  DROP INDEX `uk_student_enrollments_scope`,
  ADD COLUMN `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `student_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD COLUMN `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_student_id`,
  ADD COLUMN `active_semester_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `semester_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_academic_year_id`,
  ADD UNIQUE KEY `uk_student_enrollments_active_scope` (`active_student_id`, `active_academic_year_id`, `active_semester_id`),
  ADD KEY `idx_student_enrollments_student_id` (`student_id`);
