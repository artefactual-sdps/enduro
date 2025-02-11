-- drop foreign key constraint from "preservation_action" table
ALTER TABLE `preservation_action` DROP FOREIGN KEY `preservation_action_sip_preservation_actions`;
-- reverse: modify "sip" table
ALTER TABLE `sip` COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `id` int unsigned NOT NULL AUTO_INCREMENT,
  MODIFY COLUMN `name` varchar(2048) NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `workflow_id` varchar(255) NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `run_id` varchar(36) NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `aip_id` varchar(36) NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `location_id` varchar(36) NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `created_at` timestamp(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  MODIFY COLUMN `started_at` timestamp(6) NULL,
  MODIFY COLUMN `completed_at` timestamp(6) NULL,
  DROP INDEX `sip_aip_id_idx`,
  DROP INDEX `sip_created_at_idx`,
  DROP INDEX `sip_location_id_idx`,
  DROP INDEX `sip_name_idx`,
  DROP INDEX `sip_started_at_idx`,
  DROP INDEX `sip_status_idx`,
  ADD INDEX `package_aip_id_idx` (`aip_id`),
  ADD INDEX `package_created_at_idx` (`created_at`),
  ADD INDEX `package_location_id_idx` (`location_id`),
  ADD INDEX `package_name_idx` (`name` (50)),
  ADD INDEX `package_started_at_idx` (`started_at`),
  ADD INDEX `package_status_idx` (`status`);
-- drop foreign key constraint from "preservation_task" table
ALTER TABLE `preservation_task` DROP FOREIGN KEY `preservation_task_preservation_action_tasks`;
-- reverse: modify "preservation_action" table
ALTER TABLE `preservation_action` COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `id` int unsigned NOT NULL AUTO_INCREMENT,
  MODIFY COLUMN `workflow_id` varchar(255) NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `started_at` timestamp(6) NULL,
  MODIFY COLUMN `completed_at` timestamp(6) NULL,
  MODIFY COLUMN `sip_id` int unsigned NOT NULL,
  DROP INDEX `preservation_action_sip_preservation_actions`,
  ADD CONSTRAINT `preservation_action_ibfk_1`
    FOREIGN KEY (`sip_id`) REFERENCES `sip` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
-- reverse: modify "preservation_task" table
ALTER TABLE `preservation_task` COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `id` int unsigned NOT NULL AUTO_INCREMENT,
  MODIFY COLUMN `task_id` varchar(36) NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `name` varchar(2048) NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `started_at` timestamp(6) NULL,
  MODIFY COLUMN `completed_at` timestamp(6) NULL,
  MODIFY COLUMN `note` longtext NOT NULL COLLATE utf8mb4_0900_ai_ci,
  MODIFY COLUMN `preservation_action_id` int unsigned NOT NULL,
  DROP INDEX `preservation_task_preservation_action_tasks`,
  ADD CONSTRAINT `preservation_task_ibfk_1`
    FOREIGN KEY (`preservation_action_id`) REFERENCES `preservation_action` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
