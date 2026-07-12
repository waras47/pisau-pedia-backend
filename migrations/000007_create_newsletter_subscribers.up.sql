CREATE TABLE newsletter_subscribers (
    id         CHAR(36)     NOT NULL PRIMARY KEY,
    email      VARCHAR(255) NOT NULL,
    name       VARCHAR(150) NULL,
    source     ENUM('checkout', 'footer', 'popup', 'blog', 'manual') NOT NULL DEFAULT 'footer',
    status     ENUM('active', 'unsubscribed') NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_newsletter_subscribers_email (email),
    KEY idx_newsletter_subscribers_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
