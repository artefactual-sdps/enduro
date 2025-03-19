-- drop index and foreign key constraint from "workflow" table
ALTER TABLE `workflow` DROP INDEX `workflow_sip_workflows`, DROP FOREIGN KEY `workflow_sip_workflows`;
-- drop index and foreign key constraint from "task" table
ALTER TABLE `task` DROP INDEX `task_workflow_tasks`, DROP FOREIGN KEY `task_workflow_tasks`;
-- rename table from "workflow" to "preservation_action"
RENAME TABLE `workflow` TO `preservation_action`;
-- rename table from "task" to "preservation_task"
RENAME TABLE `task` TO `preservation_task`;
-- rename column from "workflow_id" to "preservation_action_id"
ALTER TABLE `preservation_task` CHANGE COLUMN `workflow_id` `preservation_action_id` BIGINT NOT NULL;
-- recreate index and foreign key constraint in "preservation_action" table
ALTER TABLE `preservation_action` ADD INDEX `preservation_action_sip_preservation_actions` (`sip_id`), ADD CONSTRAINT `preservation_action_sip_preservation_actions` FOREIGN KEY (`sip_id`) REFERENCES `sip` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
-- recreate index and foreign key constraint in "preservation_task" table
ALTER TABLE `preservation_task` ADD INDEX `preservation_task_preservation_action_tasks` (`preservation_action_id`), ADD CONSTRAINT `preservation_task_preservation_action_tasks` FOREIGN KEY (`preservation_action_id`) REFERENCES `preservation_action` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE;
