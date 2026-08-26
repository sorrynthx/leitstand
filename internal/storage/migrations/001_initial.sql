-- 001_initial.sql: Initial schema for Leitstand

CREATE TABLE IF NOT EXISTS hosts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL,
    address     TEXT NOT NULL,
    port        INTEGER NOT NULL DEFAULT 22,
    username    TEXT NOT NULL,
    group_name  TEXT NOT NULL DEFAULT '',
    created_at  INTEGER NOT NULL,
    updated_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS vault_meta (
    id                      INTEGER PRIMARY KEY CHECK (id = 1),
    salt                    BLOB NOT NULL,
    verification_nonce      BLOB NOT NULL,
    verification_ciphertext BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS host_secrets (
    host_id       INTEGER PRIMARY KEY,
    auth_method   TEXT NOT NULL, -- 'password', 'private_key', 'agent'
    nonce         BLOB NOT NULL,
    ciphertext    BLOB NOT NULL,
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS metrics_raw (
    host_id       INTEGER NOT NULL,
    timestamp     INTEGER NOT NULL, -- Unix timestamp in seconds
    cpu_percent   REAL NOT NULL,
    memory_total  INTEGER NOT NULL,
    memory_used   INTEGER NOT NULL,
    disk_used     INTEGER NOT NULL,
    disk_total    INTEGER NOT NULL,
    net_rx_bytes  INTEGER NOT NULL,
    net_tx_bytes  INTEGER NOT NULL,
    PRIMARY KEY (host_id, timestamp),
    FOREIGN KEY (host_id) REFERENCES hosts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_metrics_raw_timestamp
ON metrics_raw(timestamp);

CREATE INDEX IF NOT EXISTS idx_metrics_raw_host_timestamp
ON metrics_raw(host_id, timestamp);
