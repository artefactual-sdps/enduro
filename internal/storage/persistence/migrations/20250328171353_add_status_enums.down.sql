-- reverse: modify "workflow" table
ALTER TABLE `workflow` MODIFY COLUMN `status` enum('unspecified','in progress','done','error','queued','pending') NOT NULL;
-- reverse: modify "aip" table
ALTER TABLE `aip` MODIFY COLUMN `status` enum('unspecified','in_review','rejected','stored','moving') NOT NULL;
