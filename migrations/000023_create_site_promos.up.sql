CREATE TABLE site_promos (
    id              CHAR(36)     NOT NULL PRIMARY KEY,
    title           VARCHAR(255) NOT NULL,
    description     TEXT         NULL,
    discount_percent INT         NOT NULL CHECK (discount_percent BETWEEN 1 AND 100),
    popup_image     VARCHAR(500) NULL,
    start_date      DATE         NOT NULL,
    end_date        DATE         NOT NULL,
    is_active       TINYINT(1)   NOT NULL DEFAULT 1,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
