CREATE TABLE reviews (
    id             CHAR(36)     NOT NULL PRIMARY KEY,
    product_id     CHAR(36)     NOT NULL,
    customer_name  VARCHAR(150) NOT NULL,
    customer_email VARCHAR(255) NULL,
    rating         TINYINT UNSIGNED NOT NULL,
    content        TEXT         NOT NULL,
    status         ENUM('pending', 'approved', 'rejected') NOT NULL DEFAULT 'pending',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_reviews_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    CONSTRAINT chk_reviews_rating CHECK (rating BETWEEN 1 AND 5),
    KEY idx_reviews_product_id (product_id),
    KEY idx_reviews_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
