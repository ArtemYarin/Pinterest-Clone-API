CREATE TABLE pin_likes (
    user_id UUID NOT NULL,
    pin_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, pin_id)
);

CREATE INDEX idx_pin_id ON pin_likes(pin_id);