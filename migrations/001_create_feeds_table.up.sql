CREATE TABLE IF NOT EXISTS feeds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    last_fetched_at TIMESTAMP
);

CREATE INDEX idx_feeds_updated_at ON feeds (updated_at);

CREATE INDEX idx_feeds_name ON feeds (name);