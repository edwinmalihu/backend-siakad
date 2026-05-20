-- SIAKAD Core Schema
-- Target: MySQL 8.x / InnoDB / utf8mb4
-- Generated as an initial schema from the core ERD.

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` VARCHAR(100) NOT NULL,
  `password_hash` VARCHAR(255) NOT NULL,
  `email` VARCHAR(190) DEFAULT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `last_login_at` DATETIME DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_users_username` (`username`),
  UNIQUE KEY `uk_users_email` (`email`),
  KEY `idx_users_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `code` VARCHAR(100) NOT NULL,
  `description` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_roles_name` (`name`),
  UNIQUE KEY `uk_roles_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `code` VARCHAR(100) NOT NULL,
  `description` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permissions_name` (`name`),
  UNIQUE KEY `uk_permissions_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_profiles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `full_name` VARCHAR(150) NOT NULL,
  `phone` VARCHAR(30) DEFAULT NULL,
  `photo_url` VARCHAR(255) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_profiles_user_id` (`user_id`),
  CONSTRAINT `fk_user_profiles_user_id`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
    ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `user_roles` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_roles_user_role` (`user_id`, `role_id`),
  KEY `idx_user_roles_role_id` (`role_id`),
  CONSTRAINT `fk_user_roles_user_id`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT `fk_user_roles_role_id`
    FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `role_permissions` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_id` BIGINT UNSIGNED NOT NULL,
  `permission_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permissions_role_permission` (`role_id`, `permission_id`),
  KEY `idx_role_permissions_permission_id` (`permission_id`),
  CONSTRAINT `fk_role_permissions_role_id`
    FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
    ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT `fk_role_permissions_permission_id`
    FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `academic_years` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(20) NOT NULL,
  `start_date` DATE NOT NULL,
  `end_date` DATE NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_name` VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_academic_years_active_name` (`active_name`),
  KEY `idx_academic_years_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `semesters` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  `code` VARCHAR(20) NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_code` VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_semesters_active_year_code` (`active_academic_year_id`, `active_code`),
  KEY `idx_semesters_academic_year_id` (`academic_year_id`),
  KEY `idx_semesters_is_active` (`is_active`),
  CONSTRAINT `fk_semesters_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `departments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(30) NOT NULL,
  `name` VARCHAR(150) NOT NULL,
  `program_name` VARCHAR(150) DEFAULT NULL,
  `field_name` VARCHAR(150) DEFAULT NULL,
  `description` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_code` VARCHAR(30) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_name` VARCHAR(150) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_departments_active_code` (`active_code`),
  UNIQUE KEY `uk_departments_active_name` (`active_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `grade_levels` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(20) NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  `sort_order` INT NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_code` VARCHAR(20) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_name` VARCHAR(50) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_grade_levels_active_code` (`active_code`),
  UNIQUE KEY `uk_grade_levels_active_name` (`active_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `rooms` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(30) NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `type` VARCHAR(50) DEFAULT NULL,
  `capacity` INT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_code` VARCHAR(30) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_name` VARCHAR(100) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rooms_active_code` (`active_code`),
  UNIQUE KEY `uk_rooms_active_name` (`active_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `classes` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `department_id` BIGINT UNSIGNED NOT NULL,
  `grade_level_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(50) NOT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_department_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `department_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_grade_level_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `grade_level_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_name` VARCHAR(50) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_classes_active_scope` (`active_academic_year_id`, `active_department_id`, `active_grade_level_id`, `active_name`),
  KEY `idx_classes_academic_year_id` (`academic_year_id`),
  KEY `idx_classes_department_id` (`department_id`),
  KEY `idx_classes_grade_level_id` (`grade_level_id`),
  KEY `idx_classes_is_active` (`is_active`),
  CONSTRAINT `fk_classes_department_id`
    FOREIGN KEY (`department_id`) REFERENCES `departments` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_classes_grade_level_id`
    FOREIGN KEY (`grade_level_id`) REFERENCES `grade_levels` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_classes_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `students` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `nis` VARCHAR(50) NOT NULL,
  `nisn` VARCHAR(50) DEFAULT NULL,
  `full_name` VARCHAR(150) NOT NULL,
  `gender` ENUM('male', 'female') NOT NULL,
  `birth_place` VARCHAR(100) DEFAULT NULL,
  `birth_date` DATE DEFAULT NULL,
  `address` TEXT DEFAULT NULL,
  `phone` VARCHAR(30) DEFAULT NULL,
  `entry_year` YEAR NOT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'active',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_students_nis` (`nis`),
  UNIQUE KEY `uk_students_nisn` (`nisn`),
  KEY `idx_students_full_name` (`full_name`),
  KEY `idx_students_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `student_guardians` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `father_name` VARCHAR(150) DEFAULT NULL,
  `mother_name` VARCHAR(150) DEFAULT NULL,
  `guardian_name` VARCHAR(150) DEFAULT NULL,
  `guardian_phone` VARCHAR(30) DEFAULT NULL,
  `guardian_address` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_student_guardians_student_id` (`student_id`),
  CONSTRAINT `fk_student_guardians_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `teachers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `nip` VARCHAR(50) DEFAULT NULL,
  `nuptk` VARCHAR(50) DEFAULT NULL,
  `full_name` VARCHAR(150) NOT NULL,
  `gender` ENUM('male', 'female') DEFAULT NULL,
  `address` TEXT DEFAULT NULL,
  `phone` VARCHAR(30) DEFAULT NULL,
  `email` VARCHAR(190) DEFAULT NULL,
  `employment_status` VARCHAR(50) DEFAULT NULL,
  `position` VARCHAR(100) DEFAULT NULL,
  `photo_url` VARCHAR(255) DEFAULT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'active',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_teachers_nip` (`nip`),
  UNIQUE KEY `uk_teachers_nuptk` (`nuptk`),
  UNIQUE KEY `uk_teachers_email` (`email`),
  KEY `idx_teachers_full_name` (`full_name`),
  KEY `idx_teachers_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `subjects` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `department_id` BIGINT UNSIGNED NOT NULL,
  `grade_level_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(30) NOT NULL,
  `name` VARCHAR(150) NOT NULL,
  `subject_type` VARCHAR(50) DEFAULT NULL,
  `kkm` DECIMAL(5,2) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_code` VARCHAR(30) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `code`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_subjects_active_code` (`active_code`),
  KEY `idx_subjects_department_id` (`department_id`),
  KEY `idx_subjects_grade_level_id` (`grade_level_id`),
  KEY `idx_subjects_name` (`name`),
  CONSTRAINT `fk_subjects_department_id`
    FOREIGN KEY (`department_id`) REFERENCES `departments` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_subjects_grade_level_id`
    FOREIGN KEY (`grade_level_id`) REFERENCES `grade_levels` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `audit_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED DEFAULT NULL,
  `module` VARCHAR(100) NOT NULL,
  `action` VARCHAR(100) NOT NULL,
  `entity_type` VARCHAR(100) DEFAULT NULL,
  `entity_id` BIGINT UNSIGNED DEFAULT NULL,
  `payload_json` JSON DEFAULT NULL,
  `ip_address` VARCHAR(45) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_audit_logs_user_id` (`user_id`),
  KEY `idx_audit_logs_module_action` (`module`, `action`),
  KEY `idx_audit_logs_entity` (`entity_type`, `entity_id`),
  CONSTRAINT `fk_audit_logs_user_id`
    FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `announcements` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_by` BIGINT UNSIGNED DEFAULT NULL,
  `title` VARCHAR(200) NOT NULL,
  `content` TEXT NOT NULL,
  `target_scope` VARCHAR(100) DEFAULT NULL,
  `publish_start` DATETIME DEFAULT NULL,
  `publish_end` DATETIME DEFAULT NULL,
  `is_published` TINYINT(1) NOT NULL DEFAULT 0,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_announcements_created_by` (`created_by`),
  KEY `idx_announcements_target_scope` (`target_scope`),
  KEY `idx_announcements_is_published` (`is_published`),
  CONSTRAINT `fk_announcements_created_by`
    FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `student_enrollments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `class_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'active',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `student_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_semester_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `semester_id`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_student_enrollments_active_scope` (`active_student_id`, `active_academic_year_id`, `active_semester_id`),
  KEY `idx_student_enrollments_student_id` (`student_id`),
  KEY `idx_student_enrollments_class_id` (`class_id`),
  KEY `idx_student_enrollments_year_semester` (`academic_year_id`, `semester_id`),
  KEY `idx_student_enrollments_status` (`status`),
  CONSTRAINT `fk_student_enrollments_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_enrollments_class_id`
    FOREIGN KEY (`class_id`) REFERENCES `classes` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_enrollments_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_enrollments_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `student_mutations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `mutation_type` VARCHAR(30) NOT NULL,
  `from_school` VARCHAR(200) DEFAULT NULL,
  `to_school` VARCHAR(200) DEFAULT NULL,
  `reason` TEXT DEFAULT NULL,
  `effective_date` DATE DEFAULT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'pending',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_student_mutations_student_id` (`student_id`),
  KEY `idx_student_mutations_year_semester` (`academic_year_id`, `semester_id`),
  KEY `idx_student_mutations_type_status` (`mutation_type`, `status`),
  CONSTRAINT `fk_student_mutations_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_mutations_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_mutations_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `student_graduations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `graduation_date` DATE DEFAULT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'graduated',
  `notes` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `student_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_student_graduations_active_student_year` (`active_student_id`, `active_academic_year_id`),
  KEY `idx_student_graduations_student_id` (`student_id`),
  KEY `idx_student_graduations_status` (`status`),
  CONSTRAINT `fk_student_graduations_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_graduations_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `attendances` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `class_id` BIGINT UNSIGNED NOT NULL,
  `attendance_date` DATE NOT NULL,
  `status` VARCHAR(20) NOT NULL,
  `notes` TEXT DEFAULT NULL,
  `recorded_by` BIGINT UNSIGNED DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `student_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_class_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `class_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_attendance_date` DATE GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `attendance_date`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_attendances_active_student_class_date` (`active_student_id`, `active_class_id`, `active_attendance_date`),
  KEY `idx_attendances_student_id` (`student_id`),
  KEY `idx_attendances_class_id_date` (`class_id`, `attendance_date`),
  KEY `idx_attendances_recorded_by` (`recorded_by`),
  CONSTRAINT `fk_attendances_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_attendances_class_id`
    FOREIGN KEY (`class_id`) REFERENCES `classes` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_attendances_recorded_by`
    FOREIGN KEY (`recorded_by`) REFERENCES `users` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `discipline_categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL,
  `point` INT NOT NULL DEFAULT 0,
  `description` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_name` VARCHAR(100) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_discipline_categories_active_name` (`active_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `discipline_records` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `discipline_category_id` BIGINT UNSIGNED NOT NULL,
  `recorded_by` BIGINT UNSIGNED DEFAULT NULL,
  `incident_date` DATE NOT NULL,
  `description` TEXT DEFAULT NULL,
  `action_taken` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_discipline_records_student_id` (`student_id`),
  KEY `idx_discipline_records_category_id` (`discipline_category_id`),
  KEY `idx_discipline_records_recorded_by` (`recorded_by`),
  KEY `idx_discipline_records_incident_date` (`incident_date`),
  CONSTRAINT `fk_discipline_records_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_discipline_records_category_id`
    FOREIGN KEY (`discipline_category_id`) REFERENCES `discipline_categories` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_discipline_records_recorded_by`
    FOREIGN KEY (`recorded_by`) REFERENCES `users` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `extracurriculars` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `coach_teacher_id` BIGINT UNSIGNED DEFAULT NULL,
  `name` VARCHAR(150) NOT NULL,
  `description` TEXT DEFAULT NULL,
  `is_active` TINYINT(1) NOT NULL DEFAULT 1,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_name` VARCHAR(150) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_extracurriculars_active_name` (`active_name`),
  KEY `idx_extracurriculars_coach_teacher_id` (`coach_teacher_id`),
  KEY `idx_extracurriculars_is_active` (`is_active`),
  CONSTRAINT `fk_extracurriculars_coach_teacher_id`
    FOREIGN KEY (`coach_teacher_id`) REFERENCES `teachers` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `extracurricular_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `extracurricular_id` BIGINT UNSIGNED NOT NULL,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'active',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_extracurricular_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `extracurricular_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `student_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_extracurricular_members_active_scope` (`active_extracurricular_id`, `active_student_id`, `active_academic_year_id`),
  KEY `idx_extracurricular_members_extracurricular_id` (`extracurricular_id`),
  KEY `idx_extracurricular_members_student_id` (`student_id`),
  KEY `idx_extracurricular_members_academic_year_id` (`academic_year_id`),
  CONSTRAINT `fk_extracurricular_members_extracurricular_id`
    FOREIGN KEY (`extracurricular_id`) REFERENCES `extracurriculars` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_extracurricular_members_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_extracurricular_members_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `homeroom_assignments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `teacher_id` BIGINT UNSIGNED NOT NULL,
  `class_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_class_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `class_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_academic_year_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `academic_year_id`
      ELSE NULL
    END
  ) VIRTUAL,
  `active_semester_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `semester_id`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_homeroom_assignments_active_scope` (`active_class_id`, `active_academic_year_id`, `active_semester_id`),
  KEY `idx_homeroom_assignments_teacher_id` (`teacher_id`),
  KEY `idx_homeroom_assignments_class_id` (`class_id`),
  KEY `idx_homeroom_assignments_academic_year_semester` (`academic_year_id`, `semester_id`),
  CONSTRAINT `fk_homeroom_assignments_teacher_id`
    FOREIGN KEY (`teacher_id`) REFERENCES `teachers` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_homeroom_assignments_class_id`
    FOREIGN KEY (`class_id`) REFERENCES `classes` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_homeroom_assignments_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_homeroom_assignments_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `teaching_devices` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `teacher_id` BIGINT UNSIGNED NOT NULL,
  `subject_id` BIGINT UNSIGNED NOT NULL,
  `title` VARCHAR(200) NOT NULL,
  `file_url` VARCHAR(255) DEFAULT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'draft',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_teaching_devices_teacher_id` (`teacher_id`),
  KEY `idx_teaching_devices_subject_id` (`subject_id`),
  KEY `idx_teaching_devices_status` (`status`),
  CONSTRAINT `fk_teaching_devices_teacher_id`
    FOREIGN KEY (`teacher_id`) REFERENCES `teachers` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_teaching_devices_subject_id`
    FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `schedules` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `class_id` BIGINT UNSIGNED NOT NULL,
  `subject_id` BIGINT UNSIGNED NOT NULL,
  `teacher_id` BIGINT UNSIGNED NOT NULL,
  `room_id` BIGINT UNSIGNED DEFAULT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `day_of_week` TINYINT UNSIGNED DEFAULT NULL,
  `start_time` TIME DEFAULT NULL,
  `end_time` TIME DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_schedules_class_id` (`class_id`),
  KEY `idx_schedules_subject_id` (`subject_id`),
  KEY `idx_schedules_teacher_id` (`teacher_id`),
  KEY `idx_schedules_room_id` (`room_id`),
  KEY `idx_schedules_year_semester` (`academic_year_id`, `semester_id`),
  CONSTRAINT `fk_schedules_class_id`
    FOREIGN KEY (`class_id`) REFERENCES `classes` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_schedules_subject_id`
    FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_schedules_teacher_id`
    FOREIGN KEY (`teacher_id`) REFERENCES `teachers` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_schedules_room_id`
    FOREIGN KEY (`room_id`) REFERENCES `rooms` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL,
  CONSTRAINT `fk_schedules_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_schedules_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `assessment_components` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `subject_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(100) NOT NULL,
  `weight` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_assessment_components_scope` (`subject_id`, `academic_year_id`, `semester_id`, `name`),
  KEY `idx_assessment_components_year_semester` (`academic_year_id`, `semester_id`),
  CONSTRAINT `fk_assessment_components_subject_id`
    FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_assessment_components_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_assessment_components_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `student_assessments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `class_id` BIGINT UNSIGNED NOT NULL,
  `subject_id` BIGINT UNSIGNED NOT NULL,
  `assessment_component_id` BIGINT UNSIGNED NOT NULL,
  `teacher_id` BIGINT UNSIGNED NOT NULL,
  `score` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_student_assessments_scope` (`student_id`, `class_id`, `subject_id`, `assessment_component_id`, `academic_year_id`, `semester_id`),
  KEY `idx_student_assessments_teacher_id` (`teacher_id`),
  KEY `idx_student_assessments_year_semester` (`academic_year_id`, `semester_id`),
  CONSTRAINT `fk_student_assessments_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_assessments_class_id`
    FOREIGN KEY (`class_id`) REFERENCES `classes` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_assessments_subject_id`
    FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_assessments_component_id`
    FOREIGN KEY (`assessment_component_id`) REFERENCES `assessment_components` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_assessments_teacher_id`
    FOREIGN KEY (`teacher_id`) REFERENCES `teachers` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_assessments_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_assessments_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `student_grades` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `class_id` BIGINT UNSIGNED NOT NULL,
  `subject_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `semester_id` BIGINT UNSIGNED NOT NULL,
  `final_score` DECIMAL(5,2) NOT NULL DEFAULT 0.00,
  `grade_letter` VARCHAR(10) DEFAULT NULL,
  `predicate` VARCHAR(50) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_student_grades_scope` (`student_id`, `class_id`, `subject_id`, `academic_year_id`, `semester_id`),
  KEY `idx_student_grades_year_semester` (`academic_year_id`, `semester_id`),
  CONSTRAINT `fk_student_grades_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_grades_class_id`
    FOREIGN KEY (`class_id`) REFERENCES `classes` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_grades_subject_id`
    FOREIGN KEY (`subject_id`) REFERENCES `subjects` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_grades_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_student_grades_semester_id`
    FOREIGN KEY (`semester_id`) REFERENCES `semesters` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `industry_categories` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(150) NOT NULL,
  `description` TEXT DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_name` VARCHAR(150) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_industry_categories_active_name` (`active_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `companies` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `category_id` BIGINT UNSIGNED DEFAULT NULL,
  `name` VARCHAR(200) NOT NULL,
  `city` VARCHAR(100) DEFAULT NULL,
  `address` TEXT DEFAULT NULL,
  `contact_person` VARCHAR(150) DEFAULT NULL,
  `phone` VARCHAR(30) DEFAULT NULL,
  `email` VARCHAR(190) DEFAULT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'active',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  `active_name` VARCHAR(200) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_companies_active_name` (`active_name`),
  KEY `idx_companies_category_id` (`category_id`),
  KEY `idx_companies_city` (`city`),
  KEY `idx_companies_status` (`status`),
  CONSTRAINT `fk_companies_category_id`
    FOREIGN KEY (`category_id`) REFERENCES `industry_categories` (`id`)
    ON UPDATE CASCADE ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `internships` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `company_id` BIGINT UNSIGNED NOT NULL,
  `academic_year_id` BIGINT UNSIGNED NOT NULL,
  `start_date` DATE DEFAULT NULL,
  `end_date` DATE DEFAULT NULL,
  `mentor_name` VARCHAR(150) DEFAULT NULL,
  `status` VARCHAR(30) NOT NULL DEFAULT 'planned',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_internships_student_id` (`student_id`),
  KEY `idx_internships_company_id` (`company_id`),
  KEY `idx_internships_academic_year_id` (`academic_year_id`),
  KEY `idx_internships_status` (`status`),
  CONSTRAINT `fk_internships_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_internships_company_id`
    FOREIGN KEY (`company_id`) REFERENCES `companies` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT `fk_internships_academic_year_id`
    FOREIGN KEY (`academic_year_id`) REFERENCES `academic_years` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `internship_logs` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `internship_id` BIGINT UNSIGNED NOT NULL,
  `log_date` DATE NOT NULL,
  `activity` TEXT NOT NULL,
  `notes` TEXT DEFAULT NULL,
  `supervisor_name` VARCHAR(150) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_internship_logs_internship_id` (`internship_id`),
  KEY `idx_internship_logs_log_date` (`log_date`),
  CONSTRAINT `fk_internship_logs_internship_id`
    FOREIGN KEY (`internship_id`) REFERENCES `internships` (`id`)
    ON UPDATE CASCADE ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `alumni` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `student_id` BIGINT UNSIGNED NOT NULL,
  `active_student_id` BIGINT UNSIGNED GENERATED ALWAYS AS (
    CASE WHEN `deleted_at` IS NULL THEN `student_id` ELSE NULL END
  ) VIRTUAL,
  `graduation_year` YEAR NOT NULL,
  `current_activity` VARCHAR(150) DEFAULT NULL,
  `company_name` VARCHAR(200) DEFAULT NULL,
  `college_name` VARCHAR(200) DEFAULT NULL,
  `phone` VARCHAR(30) DEFAULT NULL,
  `email` VARCHAR(190) DEFAULT NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_alumni_active_student_id` (`active_student_id`),
  KEY `idx_alumni_student_id` (`student_id`),
  KEY `idx_alumni_graduation_year` (`graduation_year`),
  CONSTRAINT `fk_alumni_student_id`
    FOREIGN KEY (`student_id`) REFERENCES `students` (`id`)
    ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

SET FOREIGN_KEY_CHECKS = 1;
