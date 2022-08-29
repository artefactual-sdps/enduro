-- modify "location" table
ALTER TABLE `location` ADD COLUMN `created_at` timestamp NOT NULL;
-- modify "package" table
ALTER TABLE `package` ADD COLUMN `created_at` timestamp NOT NULL;
