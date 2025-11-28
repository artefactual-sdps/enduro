-- Add "deletion_report_key" column to "aip" table
ALTER TABLE `aip` ADD COLUMN `deletion_report_key` varchar(1024) NULL;

-- Copy existing report keys from "deletion_request" to "aip"
UPDATE `aip`,`deletion_request`
SET `aip`.`deletion_report_key` = `deletion_request`.`report_key`
WHERE `deletion_request`.`aip_id` = `aip`.`id`
AND `deletion_request`.`report_key` IS NOT NULL
AND `deletion_request`.`report_key` != '';

-- Drop "report_key" column from "deletion_request" table
ALTER TABLE `deletion_request` DROP COLUMN `report_key`;
