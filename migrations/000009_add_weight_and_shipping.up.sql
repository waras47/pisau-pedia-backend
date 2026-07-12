ALTER TABLE products
    ADD COLUMN weight INT UNSIGNED NOT NULL DEFAULT 500 AFTER stock;

ALTER TABLE orders
    ADD COLUMN shipping_cost    BIGINT UNSIGNED NOT NULL DEFAULT 0 AFTER discount_amount,
    ADD COLUMN shipping_courier VARCHAR(30)  NULL AFTER shipping_cost,
    ADD COLUMN shipping_service VARCHAR(50)  NULL AFTER shipping_courier,
    ADD COLUMN shipping_etd     VARCHAR(30)  NULL AFTER shipping_service,
    ADD COLUMN destination_id   VARCHAR(20)  NULL AFTER shipping_etd;
