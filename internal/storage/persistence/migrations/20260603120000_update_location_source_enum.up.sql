-- modify "location" table
ALTER TABLE `location` MODIFY COLUMN `source` enum('unspecified','minio','s3','sftp','amss','filesystem') NOT NULL;

-- update existing MinIO-backed locations to the broader S3 source label
UPDATE `location` SET `source` = 's3' WHERE `source` = 'minio';

-- remove the legacy MinIO source label
ALTER TABLE `location` MODIFY COLUMN `source` enum('unspecified','s3','sftp','amss','filesystem') NOT NULL;
