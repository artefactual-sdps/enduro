-- reverse: modify "sip" table
ALTER TABLE `sip` ADD COLUMN `location_id` char(36) NULL, ADD COLUMN `run_id` char(36) NOT NULL, ADD COLUMN `workflow_id` varchar(255) NOT NULL;
