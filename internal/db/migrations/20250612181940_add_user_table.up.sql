-- Create "user" table
CREATE TABLE `user` (`id` bigint NOT NULL AUTO_INCREMENT, `uuid` char(36) NOT NULL, `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, `email` varchar(1024) NULL, `name` varchar(1024) NULL, `jwt_iss` varchar(1024) NULL, `jwt_sub` varchar(1024) NULL, PRIMARY KEY (`id`), INDEX `user_jwt_iss_idx` (`jwt_iss` (50)), INDEX `user_jwt_sub_idx` (`jwt_sub` (50)), UNIQUE INDEX `uuid` (`uuid`)) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Modify "sip" table
ALTER TABLE `sip` ADD COLUMN `uploader_id` bigint NULL, ADD INDEX `sip_uploader_id_idx` (`uploader_id`), ADD CONSTRAINT `sip_user_uploaded_sips` FOREIGN KEY (`uploader_id`) REFERENCES `user` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL;
