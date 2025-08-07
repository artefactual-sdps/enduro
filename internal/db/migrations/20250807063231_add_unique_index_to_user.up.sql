-- Modify "user" table
ALTER TABLE `user` MODIFY COLUMN `oidc_iss` varchar(255) NULL, MODIFY COLUMN `oidc_sub` varchar(255) NULL, ADD UNIQUE INDEX `user_oidc_iss_sub_unique_idx` (`oidc_iss`, `oidc_sub`);
