CREATE TABLE service_requests (
    id             CHAR(36)     NOT NULL PRIMARY KEY,
    type           ENUM('sharpening', 'engraving') NOT NULL,
    status         ENUM('pending', 'in_progress', 'completed', 'rejected') NOT NULL DEFAULT 'pending',
    customer_name  VARCHAR(150) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(20)  NULL,
    message        TEXT         NOT NULL,
    quoted_price   BIGINT UNSIGNED NULL,
    admin_notes    TEXT         NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_service_requests_type (type),
    KEY idx_service_requests_status (status),
    KEY idx_service_requests_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
