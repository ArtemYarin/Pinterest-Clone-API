CREATE TABLE pins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL, -- References User Service, but NO FK constraint
    title VARCHAR(255) NOT NULL,
    image_url VARCHAR(255) NOT NULL UNIQUE,
    description VARCHAR(1000),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    likes INTEGER DEFAULT 0
);

CREATE INDEX idx_pins_title ON pins(title);
CREATE INDEX idx_pins_user_id ON pins(user_id);
CREATE INDEX idx_pins_created_at ON pins(created_at DESC);