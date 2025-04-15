-- Map "aip" statuses to new values.
UPDATE `aip` SET `status` = 'pending' WHERE `status` = 'in_review';
UPDATE `aip` SET `status` = 'processing' WHERE `status` = 'moving';
UPDATE `aip` SET `status` = 'deleted' WHERE `status` = 'rejected';

-- Update "aip" enum values.
ALTER TABLE `aip` MODIFY COLUMN `status` enum('unspecified','stored','pending','processing','deleted','queued') NOT NULL;
