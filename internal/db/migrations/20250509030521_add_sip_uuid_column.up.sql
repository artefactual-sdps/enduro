-- Modify "sip" table
ALTER TABLE `sip` ADD COLUMN `uuid` char(36) NOT NULL, ADD UNIQUE INDEX `uuid` (`uuid`);
-- Add "uuid" values
UPDATE `sip` SET `uuid` = (SELECT BIN_TO_UUID(RANDOM_BYTES(16) & 0xffffffffffff0fff3fffffffffffffff | 0x00000000000040008000000000000000)) WHERE `uuid` = NULL;
