CREATE TABLE like_counts (
    target_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    count      BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);