-- +migrate Up
-- SQL in section ''Up'' is executed when this migration is applied
CREATE TABLE `user`
(
    `id`                   int(10) unsigned NOT NULL AUTO_INCREMENT,
    `open_id`              BIGINT(20) NOT NULL DEFAULT 0,
    `username`             VARCHAR(255) NOT NULL DEFAULT '',
    `first_name`           VARCHAR(255) NOT NULL DEFAULT '',
    `last_name`            VARCHAR(255) NOT NULL DEFAULT '',
    `addr`     VARCHAR(255) NOT NULL DEFAULT '',
    `created_at`           timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`           timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`           timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_user_open_id ON user (`open_id`);
CREATE INDEX idx_user_username ON user (`username`);
CREATE INDEX idx_user_addr ON user (`addr`);

-- +migrate Down
DROP
TABL `user`;