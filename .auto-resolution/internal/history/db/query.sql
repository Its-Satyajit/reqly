-- name: InsertHistory :exec
INSERT INTO history (
    id, request_path, method, url, env, status, duration_ms, size,
    req_headers_json, req_body, req_body_path,
    resp_headers_json, resp_body, resp_body_path,
    attempts, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertHistoryFts :exec
INSERT INTO history_fts (url, request_path, id) VALUES (?, ?, ?);

-- name: ListHistory :many
SELECT id, request_path, method, url, env, status, duration_ms, size,
       req_headers_json, req_body, req_body_path,
       resp_headers_json, resp_body, resp_body_path, attempts, created_at
FROM history
ORDER BY datetime(created_at) DESC
LIMIT ? OFFSET ?;

-- name: GetHistory :one
SELECT id, request_path, method, url, env, status, duration_ms, size,
       req_headers_json, req_body, req_body_path,
       resp_headers_json, resp_body, resp_body_path, attempts, created_at
FROM history
WHERE id = ? LIMIT 1;

-- name: SearchHistoryIds :many
SELECT id FROM history_fts WHERE url MATCH ? OR request_path MATCH ? LIMIT ?;

-- name: ListHistoryIdsAsc :many
SELECT id FROM history ORDER BY datetime(created_at) ASC;

-- name: DeleteHistoryById :exec
DELETE FROM history WHERE id = ?;

-- name: DeleteHistoryFtsById :exec
DELETE FROM history_fts WHERE id = ?;

-- name: DeleteAllHistory :exec
DELETE FROM history;

-- name: DeleteAllHistoryFts :exec
DELETE FROM history_fts;

-- name: ListHistoryIdsByEnv :many
SELECT id FROM history WHERE env = ?;

-- name: UpsertCookie :exec
INSERT OR REPLACE INTO cookies (
    name, value, domain, path, expires_at, secure, http_only, same_site, env, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListCookiesByEnv :many
SELECT name, value, domain, path, expires_at, secure, http_only, same_site, env
FROM cookies WHERE env = ?;

-- name: DeleteCookie :exec
DELETE FROM cookies WHERE name = ? AND domain = ? AND path = ? AND env = ?;

-- name: DeleteAllCookies :exec
DELETE FROM cookies;

-- name: DeleteCookiesByEnv :exec
DELETE FROM cookies WHERE env = ?;
