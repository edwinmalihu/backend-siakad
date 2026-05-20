ALTER TABLE `alumni`
  ADD KEY `idx_alumni_student_id` (`student_id`),
  DROP INDEX `uk_alumni_student_id`,
  ADD COLUMN `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE WHEN `deleted_at` IS NULL THEN `student_id` ELSE NULL END
  ) VIRTUAL AFTER `student_id`,
  ADD UNIQUE KEY `uk_alumni_active_student_id` (`active_student_id`);
