CREATE TABLE coupons (
    id                    CHAR(36)     NOT NULL PRIMARY KEY,
    code                  VARCHAR(50)  NOT NULL,
    type                  ENUM('percentage', 'fixed', 'free_shipping') NOT NULL,
    value                 BIGINT UNSIGNED NOT NULL DEFAULT 0,
    min_order             BIGINT UNSIGNED NOT NULL DEFAULT 0,
    max_uses              INT UNSIGNED NULL,
    used_count            INT UNSIGNED NOT NULL DEFAULT 0,
    starts_at             DATETIME NULL,
    ends_at               DATETIME NULL,
    is_active             TINYINT(1) NOT NULL DEFAULT 1,
    description           VARCHAR(255) NULL,
    created_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_coupons_code (code),
    KEY idx_coupons_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE orders
    ADD COLUMN coupon_code VARCHAR(50) NULL AFTER currency,
    ADD COLUMN discount_amount BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER coupon_code;
