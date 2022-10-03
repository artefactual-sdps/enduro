-- modify "location" table
ALTER TABLE `location` MODIFY COLUMN `source` enum('unspecified','minio','sftp') NOT NULL;
