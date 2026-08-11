ALTER TABLE products ADD COLUMN sku VARCHAR(64) NULL AFTER slug;
ALTER TABLE products ADD UNIQUE KEY uq_products_sku (sku);
