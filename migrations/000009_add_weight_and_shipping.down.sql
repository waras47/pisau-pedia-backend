ALTER TABLE orders
    DROP COLUMN shipping_cost,
    DROP COLUMN shipping_courier,
    DROP COLUMN shipping_service,
    DROP COLUMN shipping_etd,
    DROP COLUMN destination_id;

ALTER TABLE products
    DROP COLUMN weight;
