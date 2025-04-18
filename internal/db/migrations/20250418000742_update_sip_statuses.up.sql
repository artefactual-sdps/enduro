-- Modify "sip" table to change the `status` column from an integer to an enum
-- type.
ALTER TABLE `sip` ADD COLUMN `status_str` 
    enum('error','failed','queued','processing','pending','ingested')
    NOT NULL AFTER `status`;

-- Map "new" to "error"
UPDATE `sip` SET `status_str` = 'error' WHERE `status` = 0;

-- Map "in progress" to "processing"
UPDATE `sip` SET `status_str` = 'processing' WHERE `status` = 1;

-- Map "done" to "ingested"
UPDATE `sip` SET `status_str` = 'ingested' WHERE `status` = 2;

-- Map "error" to "error"
UPDATE `sip` SET `status_str` = 'error' WHERE `status` = 3;

-- Map "unknown" to "error"
UPDATE `sip` SET `status_str` = 'error' WHERE `status` = 4;

-- Map "queued" to "queued"
UPDATE `sip` SET `status_str` = 'queued' WHERE `status` = 5;

-- Map "abandoned" to "failed"
UPDATE `sip` SET `status_str` = 'failed' WHERE `status` = 6;

-- Map "pending" to "pending"
UPDATE `sip` SET `status_str` = 'pending' WHERE `status` = 7;

ALTER TABLE `sip` DROP COLUMN `status`;
ALTER TABLE `sip` CHANGE COLUMN `status_str` `status`
    enum('error','failed','queued','processing','pending','ingested')
    NOT NULL;
