-- reverse: modify "deletion_request" table
ALTER TABLE `deletion_request` MODIFY COLUMN `status` enum('pending','approved','rejected') NOT NULL, MODIFY COLUMN `reviewer_sub` varchar(1024) NOT NULL, MODIFY COLUMN `reviewer_iss` varchar(1024) NOT NULL, MODIFY COLUMN `reviewer` varchar(1024) NOT NULL;
