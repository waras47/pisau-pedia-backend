ALTER TABLE orders
    ADD COLUMN idempotency_key VARCHAR(100) NULL AFTER id,
    ADD UNIQUE KEY uq_orders_idempotency_key (idempotency_key);
