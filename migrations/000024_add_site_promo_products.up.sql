ALTER TABLE site_promos ADD COLUMN apply_to_all TINYINT(1) NOT NULL DEFAULT 1 AFTER popup_image;

CREATE TABLE site_promo_products (
    site_promo_id CHAR(36) NOT NULL,
    product_id    CHAR(36) NOT NULL,
    PRIMARY KEY (site_promo_id, product_id),
    FOREIGN KEY (site_promo_id) REFERENCES site_promos(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
