-- reverse: modify "aip" table
ALTER TABLE `aip` DROP FOREIGN KEY `aip_location_location`, ADD CONSTRAINT `package_location_location` FOREIGN KEY (`location_id`) REFERENCES `location` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL, DROP INDEX `aip_object_key`, DROP INDEX `aip_location_location`, DROP INDEX `aip_aip_id`, ADD INDEX `pkg_object_key` (`object_key`), ADD INDEX `pkg_aip_id` (`aip_id`);
-- reverse: rename "package" table
RENAME TABLE `aip` TO `package`;
