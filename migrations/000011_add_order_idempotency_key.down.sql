ALTER TABLE orders
    DROP KEY uq_orders_idempotency_key,
    DROP COLUMN idempotency_key;
