-- modify "workflow" table
ALTER TABLE `workflow` CHANGE COLUMN `workflow_id` `temporal_id` varchar(255) NOT NULL;
