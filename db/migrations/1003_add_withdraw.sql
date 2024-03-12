-- +migrate Up
-- SQL in section ''Up'' is executed when this migration is applied
CREATE TABLE withdraw
(
    `id`          int(10) unsigned NOT NULL AUTO_INCREMENT,
    `group_id`      BIGINT(20)       NOT NULL DEFAULT 0,
    `open_id`      BIGINT(20)       NOT NULL DEFAULT 0,
    `username` VARCHAR(255) NOT NULL DEFAULT '',
    `address`     VARCHAR(255) NOT NULL DEFAULT '',
    `amount`      BIGINT(20) UNSIGNED NOT NULL DEFAULT 0,
    `tx_hash`     VARCHAR(255) NOT NULL DEFAULT '',
    `is_proceed`  tinyint(1) UNSIGNED NOT NULL DEFAULT 0,
    `created_at`  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`  timestamp NULL DEFAULT NULL,
    PRIMARY KEY (id)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
CREATE INDEX idx_withdraw_group_id ON withdraw (`group_id`);
CREATE INDEX idx_withdraw_open_id ON withdraw (`open_id`);
CREATE INDEX idx_withdraw_is_proceed ON withdraw (`is_proceed`);
-- +migrate Down
DROP
TABL `withdraw`;