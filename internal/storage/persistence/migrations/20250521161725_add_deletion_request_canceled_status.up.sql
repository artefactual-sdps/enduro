-- Modify "deletion_request" table
ALTER TABLE `deletion_request` MODIFY COLUMN `status` enum('pending','approved','rejected','canceled') NOT NULL DEFAULT "pending";
