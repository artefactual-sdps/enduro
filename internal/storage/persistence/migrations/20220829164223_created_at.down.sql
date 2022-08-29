-- reverse: modify "package" table
ALTER TABLE `package` DROP COLUMN `created_at`;
-- reverse: modify "location" table
ALTER TABLE `location` DROP COLUMN `created_at`;
