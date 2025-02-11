-- drop foreign key constraint from "preservation_action" table
ALTER TABLE `preservation_action` DROP FOREIGN KEY `preservation_action_ibfk_1`;
-- rename column from "package_id" to "sip_id"
ALTER TABLE `preservation_action` CHANGE COLUMN `package_id` `sip_id` INT UNSIGNED NOT NULL;
-- rename table from "package" to "sip"
RENAME TABLE `package` TO `sip`;
-- recreate foreign key constraint
ALTER TABLE `preservation_action` ADD CONSTRAINT `preservation_action_ibfk_1` FOREIGN KEY (`sip_id`) REFERENCES `sip` (`id`) ON DELETE CASCADE;
