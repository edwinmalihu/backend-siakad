ALTER TABLE `discipline_categories`
  DROP INDEX `uk_discipline_categories_name`,
  ADD COLUMN `active_name` VARCHAR(100) GENERATED ALWAYS AS (
    CASE
      WHEN `deleted_at` IS NULL THEN `name`
      ELSE NULL
    END
  ) VIRTUAL AFTER `deleted_at`,
  ADD UNIQUE KEY `uk_discipline_categories_active_name` (`active_name`);
