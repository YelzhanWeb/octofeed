CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    title TEXT NOT NULL,
    link TEXT NOT NULL,
    published_at TIMESTAMP,
    description TEXT,
    feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
    UNIQUE (link, feed_id)
);

CREATE INDEX idx_articles_feed_id ON articles (feed_id);

CREATE INDEX idx_articles_published_at ON articles (published_at DESC);

CREATE INDEX idx_articles_created_at ON articles (created_at DESC);