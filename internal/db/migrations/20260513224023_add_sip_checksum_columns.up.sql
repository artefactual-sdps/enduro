-- Modify "sip" table
ALTER TABLE `sip` ADD COLUMN `checksum_algorithm` varchar(255) NULL, ADD COLUMN `checksum_value` varchar(255) NULL, ADD INDEX `sip_checksum_value_idx` (`checksum_value`);
