-- drop foreign key constraint from "preservation_action" table
ALTER TABLE `preservation_action` DROP FOREIGN KEY `preservation_action_ibfk_1`;
-- reverse: rename column from "package_id" to "sip_id"
ALTER TABLE `preservation_action` CHANGE COLUMN `sip_id` `package_id` INT UNSIGNED NOT NULL;
-- reverse: rename table from "package" to "sip"
RENAME TABLE `sip` TO `package`;
-- recreate foreign key constraint
ALTER TABLE `preservation_action` ADD CONSTRAINT `preservation_action_ibfk_1` FOREIGN KEY (`package_id`) REFERENCES `package` (`id`) ON DELETE CASCADE;
