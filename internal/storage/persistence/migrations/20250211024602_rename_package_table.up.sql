-- rename "package" table
RENAME TABLE `package` TO `aip`;
-- modify "aip" table
ALTER TABLE `aip` DROP INDEX `pkg_aip_id`, DROP INDEX `pkg_object_key`, ADD INDEX `aip_aip_id` (`aip_id`), ADD INDEX `aip_location_location` (`location_id`), ADD INDEX `aip_object_key` (`object_key`), DROP FOREIGN KEY `package_location_location`, ADD CONSTRAINT `aip_location_location` FOREIGN KEY (`location_id`) REFERENCES `location` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL;
