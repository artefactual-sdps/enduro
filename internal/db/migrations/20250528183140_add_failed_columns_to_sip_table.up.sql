-- Modify "sip" table
ALTER TABLE `sip` ADD COLUMN `failed_as` enum('SIP','PIP') NULL, ADD COLUMN `failed_key` varchar(1024) NULL;
