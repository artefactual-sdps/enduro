-- modify "aip" table
ALTER TABLE `aip` MODIFY COLUMN `status` enum('unspecified','in_review','rejected','stored','moving','pending','processing','deleted') NOT NULL;
-- modify "workflow" table
ALTER TABLE `workflow` MODIFY COLUMN `status` enum('unspecified','in progress','done','error','queued','pending','canceled') NOT NULL;
