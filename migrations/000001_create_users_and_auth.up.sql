CREATE TABLE users (
    id             CHAR(36)     NOT NULL PRIMARY KEY,
    email          VARCHAR(255) NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    full_name      VARCHAR(150) NOT NULL,
    phone          VARCHAR(20)  NULL,
    role           ENUM('customer', 'admin') NOT NULL DEFAULT 'customer',
    avatar_url     VARCHAR(500) NULL,
    is_active      TINYINT(1)   NOT NULL DEFAULT 1,
    email_verified_at DATETIME NULL,
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE addresses (
    id           CHAR(36)     NOT NULL PRIMARY KEY,
    user_id      CHAR(36)     NOT NULL,
    label        VARCHAR(50)  NULL,
    full_name    VARCHAR(150) NOT NULL,
    phone        VARCHAR(20)  NULL,
    address_line VARCHAR(500) NOT NULL,
    city         VARCHAR(100) NOT NULL,
    province     VARCHAR(100) NULL,
    postal_code  VARCHAR(10)  NOT NULL,
    is_default   TINYINT(1)   NOT NULL DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    KEY idx_addresses_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE refresh_tokens (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    user_id     CHAR(36)     NOT NULL,
    token_hash  VARCHAR(255) NOT NULL,
    expires_at  DATETIME     NOT NULL,
    revoked_at  DATETIME     NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    KEY idx_refresh_tokens_user_id (user_id),
    UNIQUE KEY uq_refresh_tokens_hash (token_hash)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
