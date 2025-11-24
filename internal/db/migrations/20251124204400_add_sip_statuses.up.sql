-- Modify "sip" table
ALTER TABLE `sip` MODIFY COLUMN `status` enum('error','failed','queued','processing','pending','ingested','validated','canceled') NOT NULL;
