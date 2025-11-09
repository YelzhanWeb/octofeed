CREATE TABLE IF NOT EXISTS ipc_commands (
    command TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS aggregator_config (
    id INTEGER PRIMARY KEY DEFAULT 1,
    interval_seconds INTEGER NOT NULL DEFAULT 180,
    workers_count INTEGER NOT NULL DEFAULT 3,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (id = 1) -- Только одна запись
);

INSERT INTO
    aggregator_config (
        id,
        interval_seconds,
        workers_count
    )
VALUES (1, 180, 3)
ON CONFLICT (id) DO NOTHING;