ALTER TABLE `student_graduations`
  DROP INDEX `uk_student_graduations_student_year`,
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
  ADD UNIQUE KEY `uk_student_graduations_active_student_year` (`active_student_id`, `active_academic_year_id`),
  ADD KEY `idx_student_graduations_student_id` (`student_id`);

ALTER TABLE `attendances`
  DROP INDEX `uk_attendances_student_class_date`,
  ADD COLUMN `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `student_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD COLUMN `active_class_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `class_id`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_student_id`,
  ADD COLUMN `active_attendance_date` DATE GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `attendance_date`
      ELSE NULL
    END
  ) VIRTUAL AFTER `active_class_id`,
  ADD UNIQUE KEY `uk_attendances_active_student_class_date` (`active_student_id`, `active_class_id`, `active_attendance_date`),
  ADD KEY `idx_attendances_student_id` (`student_id`);
