CREATE TABLE storage_package (
  `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
  `name` VARCHAR(2048) NOT NULL,
  `aip_id` VARCHAR(36) NOT NULL,
  `location` VARCHAR(2048) NOT NULL,
  `status` TINYINT NOT NULL, -- {in_review, rejected, stored}
  `object_key` VARCHAR(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `package_aip_id_idx` (`aip_id`),
  KEY `package_location_idx` (`location`(50)),
  KEY `package_object_key_idx` (`object_key`)
);
