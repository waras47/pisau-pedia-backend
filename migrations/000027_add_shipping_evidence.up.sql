ALTER TABLE orders
    ADD COLUMN tracking_number VARCHAR(100) DEFAULT NULL AFTER shipping_etd,
    ADD COLUMN shipping_evidence_url VARCHAR(500) DEFAULT NULL AFTER tracking_number;
