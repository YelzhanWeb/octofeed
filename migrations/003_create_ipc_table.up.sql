CREATE TABLE IF NOT EXISTS ipc_commands (
    command TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO
    aggregator_config (
        id,
        interval_seconds,
        workers_count
    )
VALUES (1, 180, 3)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS aggregator_lock (
    id INTEGER PRIMARY KEY DEFAULT 1,
    lock_id TEXT NOT NULL,
    locked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (id = 1)
);