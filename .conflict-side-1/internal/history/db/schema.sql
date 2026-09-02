-- Canonical snapshot of the history database schema for sqlc code
-- generation. Runtime migrations in history.go remain the source of truth
-- for fresh databases and legacy upgrades; keep this file in sync.

CREATE TABLE IF NOT EXISTS history (
    id TEXT PRIMARY KEY,
    request_path TEXT,
    method TEXT,
    url TEXT,
    env TEXT,
    status INTEGER,
    duration_ms INTEGER,
    size INTEGER,
    req_headers_json TEXT,
    req_body BLOB,
    req_body_path TEXT,
    resp_headers_json TEXT,
    resp_body BLOB,
    resp_body_path TEXT,
    attempts INTEGER DEFAULT 1,
    created_at TEXT NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS history_fts USING fts5(
    url,
    request_path,
    id UNINDEXED
);

CREATE TABLE IF NOT EXISTS cookies (
    name TEXT,
    value TEXT,
    domain TEXT,
    path TEXT,
    expires_at TEXT NOT NULL,
    secure INTEGER,
    http_only INTEGER,
    same_site TEXT,
    env TEXT,
    created_at TEXT NOT NULL,
    UNIQUE(name, domain, path, env)
);

CREATE INDEX IF NOT EXISTS idx_history_created ON history(created_at);
CREATE INDEX IF NOT EXISTS idx_history_status ON history(status);
CREATE INDEX IF NOT EXISTS idx_history_env ON history(env);
