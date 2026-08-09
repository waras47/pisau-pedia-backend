CREATE TABLE post_categories (
    id          CHAR(36)     NOT NULL PRIMARY KEY,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    name_id     VARCHAR(255) NOT NULL,
    name_en     VARCHAR(255) NOT NULL,
    desc_id     TEXT         NULL,
    desc_en     TEXT         NULL,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE posts (
    id              CHAR(36)     NOT NULL PRIMARY KEY,
    slug            VARCHAR(255) NOT NULL UNIQUE,
    category_id     CHAR(36)     NOT NULL,
    title_id        VARCHAR(500) NOT NULL,
    title_en        VARCHAR(500) NOT NULL,
    excerpt_id      TEXT         NULL,
    excerpt_en      TEXT         NULL,
    content_id      LONGTEXT     NULL,
    content_en      LONGTEXT     NULL,
    image           VARCHAR(500) NULL,
    reading_minutes INT          NOT NULL DEFAULT 5,
    status          ENUM('draft','published') NOT NULL DEFAULT 'draft',
    published_at    DATETIME     NULL,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES post_categories(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
