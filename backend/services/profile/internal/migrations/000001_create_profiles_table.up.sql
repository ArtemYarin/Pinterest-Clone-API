CREATE TABLE profiles (
    user_id UUID PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    bio VARCHAR(500) NOT NULL DEFAULT '...',
    avatar_key VARCHAR(255),
    avatar_status VARCHAR(20) NOT NULL DEFAULT 'none'
        CHECK (avatar_status IN ('none', 'pending', 'confirmed')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_profiles_username ON profiles(username);
