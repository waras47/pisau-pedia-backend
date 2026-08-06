CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id         CHAR(36) NOT NULL PRIMARY KEY,
    user_id    CHAR(36) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_verification_token_hash (token_hash),
    KEY idx_verification_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

UPDATE users SET email_verified_at = NOW() WHERE role = 'admin' AND email_verified_at IS NULL;
