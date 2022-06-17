CREATE TABLE storage_package (
  `id` INT UNSIGNED NOT NULL,
  `name` VARCHAR(2048) NOT NULL,
  `status` TINYINT NOT NULL, -- {review, rejected, permanent}
  `key` VARCHAR(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `package_key_idx` (`key`)
);
