CREATE INDEX idx_pins_pending_created_at ON pins(created_at) WHERE image_status = 'pending';
