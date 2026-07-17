ALTER TABLE users
    ADD COLUMN google_id VARCHAR(255) NULL AFTER password_hash,
    ADD UNIQUE KEY uq_users_google_id (google_id);
