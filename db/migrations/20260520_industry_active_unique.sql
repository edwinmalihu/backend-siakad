ALTER TABLE `industry_categories`
  DROP INDEX `uk_industry_categories_name`,
  ADD COLUMN `active_name` VARCHAR(150)
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL THEN `name`
        ELSE NULL
      END
    ) VIRTUAL
    AFTER `deleted_at`,
  ADD UNIQUE KEY `uk_industry_categories_active_name` (`active_name`);

ALTER TABLE `companies`
  DROP INDEX `uk_companies_name`,
  ADD COLUMN `active_name` VARCHAR(200)
    GENERATED ALWAYS AS (
      CASE
        WHEN `deleted_at` IS NULL THEN `name`
        ELSE NULL
      END
    ) VIRTUAL
    AFTER `deleted_at`,
  ADD UNIQUE KEY `uk_companies_active_name` (`active_name`);
