-- Add "type_str" enum column to "workflow" table.
ALTER TABLE `workflow` ADD COLUMN `type_str` enum('create aip','create and review aip') NOT NULL AFTER `type`;

-- Delete "unspecified" type workflows.
DELETE FROM `workflow` WHERE `type` = 0;

-- Update "type_str" for "create aip" workflows.
UPDATE `workflow` SET `type_str` = 'create aip' WHERE `type` = 1;

-- Update "type_str" for "create and review aip" workflows.
UPDATE `workflow` SET `type_str` = 'create and review aip' WHERE `type` = 2;

-- Drop "type" column from "workflow" table.
ALTER TABLE `workflow` DROP COLUMN `type`;

-- Rename "type_str" column to "type" in "workflow" table.
ALTER TABLE `workflow` CHANGE COLUMN `type_str` `type` enum('create aip','create and review aip') NOT NULL;
