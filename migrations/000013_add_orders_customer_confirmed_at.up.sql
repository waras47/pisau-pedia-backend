ALTER TABLE orders
    ADD COLUMN customer_confirmed_at DATETIME NULL AFTER payment_status;
