ALTER TABLE users
    DROP KEY uq_users_google_id,
    DROP COLUMN google_id;
