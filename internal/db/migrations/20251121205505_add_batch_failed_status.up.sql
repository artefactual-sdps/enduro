-- Modify "batch" table
ALTER TABLE `batch` MODIFY COLUMN `status` enum('queued','processing','pending','ingested','canceled','failed') NOT NULL;
