CREATE TABLE package (
  `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
  `name` VARCHAR(2048) NOT NULL,
  `workflow_id` VARCHAR(255) NOT NULL,
  `run_id` VARCHAR(36) NOT NULL,
  `aip_id` VARCHAR(36) NOT NULL,
  `location` VARCHAR(2048) NOT NULL,
  `status` TINYINT NOT NULL,
  `created_at` TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) NOT NULL,
  `started_at` TIMESTAMP(6) NULL,
  `completed_at` TIMESTAMP(6) NULL,
  PRIMARY KEY (`id`),
  KEY `package_name_idx` (`name`(50)),
  KEY `package_aip_id_idx` (`aip_id`),
  KEY `package_location_idx` (`location`(50)),
  KEY `package_status_idx` (`status`),
  KEY `package_created_at_idx` (`created_at`),
  KEY `package_started_at_idx` (`started_at`)
);
CREATE TABLE preservation_action (
  `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
  `workflow_id` VARCHAR(255) NOT NULL,
  `type` TINYINT NOT NULL,
  `status` TINYINT NOT NULL,
  `started_at` TIMESTAMP(6) NULL,
  `completed_at` TIMESTAMP(6) NULL,
  `package_id` INT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  FOREIGN KEY (`package_id`) REFERENCES package(`id`) ON DELETE CASCADE
);
CREATE TABLE preservation_task (
  `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
  `task_id` VARCHAR(36) NOT NULL,
  `name` VARCHAR(2048) NOT NULL,
  `status` TINYINT NOT NULL,
  `started_at` TIMESTAMP(6) NULL,
  `completed_at` TIMESTAMP(6) NULL,
  `note` LONGTEXT NOT NULL,
  `preservation_action_id` INT UNSIGNED NOT NULL,
  PRIMARY KEY (`id`),
  FOREIGN KEY (`preservation_action_id`) REFERENCES preservation_action(`id`) ON DELETE CASCADE
);
