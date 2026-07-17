ALTER TABLE product_images
    ADD COLUMN angle ENUM('front', 'back', 'side', 'top') NULL AFTER alt_text;

ALTER TABLE products
    ADD COLUMN care_instructions TEXT NULL AFTER description;
