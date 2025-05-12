-- Add nullable "uuid" column to "sip" table
ALTER TABLE `sip` ADD COLUMN `uuid` CHAR(36) NULL;
-- Add "uuid" values
UPDATE `sip` SET `uuid` = BIN_TO_UUID(
    RANDOM_BYTES(16) & 0xffffffffffff0fff3fffffffffffffff | 0x00000000000040008000000000000000
) WHERE uuid IS NULL;
-- Make "uuid" column not nullable and add unique index
ALTER TABLE `sip` MODIFY COLUMN `uuid` CHAR(36) NOT NULL, ADD UNIQUE INDEX `uuid` (`uuid`);
