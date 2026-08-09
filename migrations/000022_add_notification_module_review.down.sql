ALTER TABLE notifications MODIFY COLUMN module ENUM('product', 'order', 'customer', 'service') NOT NULL;
