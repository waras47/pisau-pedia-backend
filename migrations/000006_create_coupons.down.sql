ALTER TABLE orders
    DROP COLUMN discount_amount,
    DROP COLUMN coupon_code;

DROP TABLE IF EXISTS coupons;
