ALTER TABLE pins ADD COLUMN image_status VARCHAR(20) NOT NULL DEFAULT 'pending';
ALTER TABLE pins ADD CONSTRAINT pins_image_status_check CHECK (image_status IN ('pending', 'confirmed', 'failed'));
