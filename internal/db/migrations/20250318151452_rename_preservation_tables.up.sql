-- drop index and foreign key constraint from "preservation_action" table
ALTER TABLE `preservation_action` DROP INDEX `preservation_action_sip_preservation_actions`, DROP FOREIGN KEY `preservation_action_sip_preservation_actions`;
-- drop index and foreign key constraint from "preservation_task" table
ALTER TABLE `preservation_task` DROP INDEX `preservation_task_preservation_action_tasks`, DROP FOREIGN KEY `preservation_task_preservation_action_tasks`;
-- rename table from "preservation_action" to "workflow"
RENAME TABLE `preservation_action` TO `workflow`;
-- rename table from "preservation_task" to "task"
RENAME TABLE `preservation_task` TO `task`;
-- rename column from "preservation_action_id" to "workflow_id"
ALTER TABLE `task` CHANGE COLUMN `preservation_action_id` `workflow_id` BIGINT NOT NULL;
-- recreate index and foreign key constraint in "workflow" table
ALTER TABLE `workflow` ADD INDEX `workflow_sip_workflows` (`sip_id`), ADD CONSTRAINT `workflow_sip_workflows` FOREIGN KEY (`sip_id`) REFERENCES `sip` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
-- recreate index and foreign key constraint in "task" table
ALTER TABLE `task` ADD INDEX `task_workflow_tasks` (`workflow_id`), ADD CONSTRAINT `task_workflow_tasks` FOREIGN KEY (`workflow_id`) REFERENCES `workflow` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
