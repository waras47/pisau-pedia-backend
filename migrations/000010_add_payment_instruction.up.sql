ALTER TABLE orders
    ADD COLUMN payment_provider VARCHAR(20)  NULL AFTER payment_status,
    ADD COLUMN payment_id       VARCHAR(100) NULL AFTER payment_provider,
    ADD COLUMN payment_type     VARCHAR(30)  NULL AFTER payment_id,
    ADD COLUMN payment_channel  VARCHAR(30)  NULL AFTER payment_type,
    ADD COLUMN payment_va_number VARCHAR(50) NULL AFTER payment_channel,
    ADD COLUMN payment_qr_string TEXT        NULL AFTER payment_va_number,
    ADD COLUMN payment_url      VARCHAR(500) NULL AFTER payment_qr_string,
    ADD COLUMN payment_expiry   VARCHAR(40)  NULL AFTER payment_url;
