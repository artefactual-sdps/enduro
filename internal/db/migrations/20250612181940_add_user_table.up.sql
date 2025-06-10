-- Create "user" table
CREATE TABLE `user` (`id` bigint NOT NULL AUTO_INCREMENT, `uuid` char(36) NOT NULL, `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP, `email` varchar(1024) NULL, `name` varchar(1024) NULL, `oidc_iss` varchar(1024) NULL, `oidc_sub` varchar(1024) NULL, PRIMARY KEY (`id`), INDEX `user_oidc_iss_idx` (`oidc_iss` (50)), INDEX `user_oidc_sub_idx` (`oidc_sub` (50)), UNIQUE INDEX `uuid` (`uuid`)) CHARSET utf8mb4 COLLATE utf8mb4_bin;
-- Modify "sip" table
ALTER TABLE `sip` ADD COLUMN `uploader_id` bigint NULL, ADD INDEX `sip_uploader_id_idx` (`uploader_id`), ADD CONSTRAINT `sip_user_uploaded_sips` FOREIGN KEY (`uploader_id`) REFERENCES `user` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL;
