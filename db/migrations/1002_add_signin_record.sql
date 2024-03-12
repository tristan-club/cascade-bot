-- +migrate Up
-- SQL in section ''Up'' is executed when this migration is applied
CREATE TABLE signin_record
(
    `id`                    int(10) unsigned NOT NULL AUTO_INCREMENT,
    `group_id`              BIGINT(20) NOT NULL DEFAULT 0,
    `open_id`               BIGINT(20) NOT NULL DEFAULT 0,
    `reward`                int(20) NOT NULL DEFAULT 0,
    `config_id`             BIGINT(20) NOT NULL DEFAULT 0,
    `extra_reward`          BIGINT(20) NOT NULL DEFAULT 0,
    `created_at`            timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`            timestamp     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`            timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_signin_record_group_id ON signin_record (`group_id`);
CREATE INDEX idx_signin_record_open_id ON signin_record (`open_id`);
CREATE INDEX idx_signin_record_created_at ON signin_record (`created_at`);
-- +migrate Down
DROP
TABL `signin_record`;