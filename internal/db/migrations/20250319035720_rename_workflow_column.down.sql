-- reverse: modify "workflow" table
ALTER TABLE `workflow` CHANGE COLUMN `temporal_id` `workflow_id` varchar(255) NOT NULL;
