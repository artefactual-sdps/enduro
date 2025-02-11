-- drop foreign key constraint from "preservation_action" table
ALTER TABLE `preservation_action` DROP FOREIGN KEY `preservation_action_ibfk_1`;
-- modify "sip" table
ALTER TABLE `sip` COLLATE utf8mb4_bin,
  MODIFY COLUMN `id` bigint NOT NULL AUTO_INCREMENT,
  MODIFY COLUMN `name` varchar(2048) NOT NULL,
  MODIFY COLUMN `workflow_id` varchar(255) NOT NULL,
  MODIFY COLUMN `run_id` char(36) NOT NULL,
  MODIFY COLUMN `aip_id` char(36) NULL,
  MODIFY COLUMN `location_id` char(36) NULL,
  MODIFY COLUMN `created_at` timestamp NOT NULL,
  MODIFY COLUMN `started_at` timestamp NULL,
  MODIFY COLUMN `completed_at` timestamp NULL,
  DROP INDEX `package_aip_id_idx`,
  DROP INDEX `package_created_at_idx`,
  DROP INDEX `package_location_id_idx`,
  DROP INDEX `package_name_idx`,
  DROP INDEX `package_started_at_idx`,
  DROP INDEX `package_status_idx`,
  ADD INDEX `sip_aip_id_idx` (`aip_id`),
  ADD INDEX `sip_created_at_idx` (`created_at`),
  ADD INDEX `sip_location_id_idx` (`location_id`),
  ADD INDEX `sip_name_idx` (`name` (50)),
  ADD INDEX `sip_started_at_idx` (`started_at`),
  ADD INDEX `sip_status_idx` (`status`);
-- drop foreign key constraint from "preservation_task" table
ALTER TABLE `preservation_task` DROP FOREIGN KEY `preservation_task_ibfk_1`;
-- modify "preservation_action" table
ALTER TABLE `preservation_action` COLLATE utf8mb4_bin,
  MODIFY COLUMN `id` bigint NOT NULL AUTO_INCREMENT,
  MODIFY COLUMN `workflow_id` varchar(255) NOT NULL,
  MODIFY COLUMN `started_at` timestamp NULL,
  MODIFY COLUMN `completed_at` timestamp NULL,
  MODIFY COLUMN `sip_id` bigint NOT NULL,
  ADD INDEX `preservation_action_sip_preservation_actions` (`sip_id`),
  ADD CONSTRAINT `preservation_action_sip_preservation_actions`
    FOREIGN KEY (`sip_id`) REFERENCES `sip` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "preservation_task" table
ALTER TABLE `preservation_task` COLLATE utf8mb4_bin,
  MODIFY COLUMN `id` bigint NOT NULL AUTO_INCREMENT,
  MODIFY COLUMN `task_id` char(36) NOT NULL,
  MODIFY COLUMN `name` varchar(2048) NOT NULL,
  MODIFY COLUMN `started_at` timestamp NULL,
  MODIFY COLUMN `completed_at` timestamp NULL,
  MODIFY COLUMN `note` longtext NOT NULL,
  MODIFY COLUMN `preservation_action_id` bigint NOT NULL,
  ADD INDEX `preservation_task_preservation_action_tasks` (`preservation_action_id`),
  ADD CONSTRAINT `preservation_task_preservation_action_tasks`
    FOREIGN KEY (`preservation_action_id`) REFERENCES `preservation_action` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
