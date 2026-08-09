ALTER TABLE notifications MODIFY COLUMN module ENUM('product', 'order', 'customer', 'service', 'review') NOT NULL;
