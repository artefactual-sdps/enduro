-- Rename "task_id" column to "uuid" in "task" table and add unique index
ALTER TABLE `task` CHANGE COLUMN `task_id` `uuid` CHAR(36) NOT NULL, ADD UNIQUE INDEX `uuid` (`uuid`);
-- Add nullable "uuid" column to "workflow" table
ALTER TABLE `workflow` ADD COLUMN `uuid` CHAR(36) NULL;
-- Add "uuid" values
UPDATE `workflow` SET `uuid` = BIN_TO_UUID(
    RANDOM_BYTES(16) & 0xffffffffffff0fff3fffffffffffffff | 0x00000000000040008000000000000000
) WHERE uuid IS NULL;
-- Make "uuid" column not nullable and add unique index
ALTER TABLE `workflow` MODIFY COLUMN `uuid` CHAR(36) NOT NULL, ADD UNIQUE INDEX `uuid` (`uuid`);
