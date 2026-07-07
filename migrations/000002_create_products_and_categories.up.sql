CREATE TABLE categories (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(120) NOT NULL,
    description TEXT NULL,
    image_url   VARCHAR(500) NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE products (
    id                CHAR(36)     NOT NULL PRIMARY KEY,
    category_id       CHAR(36)     NULL,
    name              VARCHAR(200) NOT NULL,
    slug              VARCHAR(220) NOT NULL,
    description       TEXT NULL,
    price             BIGINT UNSIGNED NOT NULL,
    compare_at_price  BIGINT UNSIGNED NULL,
    currency          VARCHAR(3)   NOT NULL DEFAULT 'IDR',
    maker             VARCHAR(150) NULL,
    badge             ENUM('new', 'sale', 'sold-out') NULL,
    stock             INT UNSIGNED NOT NULL DEFAULT 0,
    rating_avg        DECIMAL(2,1) NOT NULL DEFAULT 0.0,
    review_count      INT UNSIGNED NOT NULL DEFAULT 0,
    is_active         TINYINT(1)   NOT NULL DEFAULT 1,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_products_slug (slug),
    CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
    KEY idx_products_category_id (category_id),
    KEY idx_products_is_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_images (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    url         VARCHAR(500) NOT NULL,
    alt_text    VARCHAR(200) NULL,
    sort_order  INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    KEY idx_product_images_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_specs (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    label       VARCHAR(100) NOT NULL,
    value       VARCHAR(200) NOT NULL,
    sort_order  INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_product_specs_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    KEY idx_product_specs_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE product_highlights (
    id          CHAR(36) NOT NULL PRIMARY KEY,
    product_id  CHAR(36) NOT NULL,
    highlight   VARCHAR(255) NOT NULL,
    sort_order  INT UNSIGNED NOT NULL DEFAULT 0,
    CONSTRAINT fk_product_highlights_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    KEY idx_product_highlights_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
