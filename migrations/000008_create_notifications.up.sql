CREATE TABLE notifications (
    id           CHAR(36)     NOT NULL PRIMARY KEY,
    module       ENUM('product', 'order', 'customer', 'service') NOT NULL,
    type         VARCHAR(50)  NOT NULL,
    title        VARCHAR(255) NOT NULL,
    message      TEXT         NOT NULL,
    reference_id CHAR(36)     NULL,
    link         VARCHAR(255) NULL,
    is_read      TINYINT(1)   NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_notifications_created_at (created_at),
    KEY idx_notifications_is_read (is_read)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
