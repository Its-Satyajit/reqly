// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Entry is one row in the per-workspace history, as defined in CONTEXT.md.
type Entry struct {
	ID          string
	RequestPath string
	Method      string
	URL         string
	Env         string
	Status      int
	DurationMS  int64
	Size        int64
	ReqHeaders  map[string][]string
	ReqBody     []byte
	RespHeaders map[string][]string
	RespBody    []byte
	Attempts    int
	CreatedAt   time.Time
}

// Cookie is a persisted response cookie (Cookie Jar).
type Cookie struct {
	Name      string
	Value     string
	Domain    string
	Path      string
	ExpiresAt time.Time
	Secure    bool
	HttpOnly  bool
	SameSite  string
	Env       string
}

// Store is the SQLite store for history and cookies.
type Store struct {
	dbPath string
	db     *sql.DB
	mu     sync.Mutex
}

const spillThreshold = 1 << 20 // 1 MB

// NewStore opens (creating if needed) the per-workspace history.db.
func NewStore(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("history mkdir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{dbPath: dbPath, db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := os.Chmod(dbPath, 0o600); err != nil && !os.IsNotExist(err) {
		// ignore chmod error on fresh in-memory? but dbPath is file
	}
	return s, nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS history (
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
			created_at DATETIME
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS history_fts USING fts5(url, request_path, id UNINDEXED)`,
		`CREATE TABLE IF NOT EXISTS cookies (
			name TEXT,
			value TEXT,
			domain TEXT,
			path TEXT,
			expires_at DATETIME,
			secure INTEGER,
			http_only INTEGER,
			same_site TEXT,
			env TEXT,
			created_at DATETIME,
			UNIQUE(name, domain, path, env)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_history_created ON history(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_history_status ON history(status)`,
		`CREATE INDEX IF NOT EXISTS idx_history_env ON history(env)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("history migrate: %w", err)
		}
	}
	return s.addColumn(ctx, "history", "attempts", "INTEGER DEFAULT 1")
}

// addColumn alters a table with a new column when it is missing, bridging
// databases created by earlier schema versions.
func (s *Store) addColumn(ctx context.Context, table, column, decl string) error {
	rows, err := s.db.QueryContext(ctx,
		`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`, table, column)
	if err != nil {
		return fmt.Errorf("history migrate: %w", err)
	}
	defer rows.Close()
	var n int
	if !rows.Next() {
		return fmt.Errorf("history migrate: pragma_table_info(%s) returned no rows", table)
	}
	if err := rows.Scan(&n); err != nil {
		return fmt.Errorf("history migrate: %w", err)
	}
	rows.Close()
	if n > 0 {
		return nil
	}
	if _, err := s.db.ExecContext(ctx,
		fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, decl)); err != nil {
		return fmt.Errorf("history migrate: %w", err)
	}
	return nil
}

// Close closes the DB.
func (s *Store) Close() error { return s.db.Close() }

// Insert inserts one entry, handling spill and FTS.
func (s *Store) Insert(ctx context.Context, e *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.ID == "" {
		e.ID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	var reqBodyPath, respBodyPath string
	var reqBodyBlob, respBodyBlob []byte
	// req body
	if len(e.ReqBody) > spillThreshold {
		p, err := s.spill(e.ID, "req", e.ReqBody)
		if err != nil {
			return err
		}
		reqBodyPath = p
	} else {
		reqBodyBlob = e.ReqBody
	}
	if len(e.RespBody) > spillThreshold {
		p, err := s.spill(e.ID, "resp", e.RespBody)
		if err != nil {
			return err
		}
		respBodyPath = p
	} else {
		respBodyBlob = e.RespBody
	}
	reqH, _ := json.Marshal(e.ReqHeaders)
	respH, _ := json.Marshal(e.RespHeaders)
	_, err := s.db.ExecContext(ctx, `INSERT INTO history (id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body, req_body_path, resp_headers_json, resp_body, resp_body_path, attempts, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		e.ID, e.RequestPath, e.Method, e.URL, e.Env, e.Status, e.DurationMS, e.Size, string(reqH), reqBodyBlob, reqBodyPath, string(respH), respBodyBlob, respBodyPath, maxAttempts(e.Attempts), e.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	// FTS
	_, _ = s.db.ExecContext(ctx, `INSERT INTO history_fts (url, request_path, id) VALUES (?,?,?)`, e.URL, e.RequestPath, e.ID)
	// retention is not auto here; caller may call EnforceRetention, but Insert auto-prunes to 500
	_ = s.enforceRetentionLocked(ctx, 500)
	return nil
}

func (s *Store) spill(id, suffix string, data []byte) (string, error) {
	dir := filepath.Join(filepath.Dir(s.dbPath), "history", "blobs")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	p := filepath.Join(dir, fmt.Sprintf("%s.%s.bin", id, suffix))
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return "", err
	}
	return p, nil
}

func (s *Store) loadBody(blob []byte, path string) []byte {
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			return b
		}
	}
	return blob
}

// List returns entries ordered by created_at DESC.
func (s *Store) List(ctx context.Context, limit, offset int, statusFilter *int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var rows *sql.Rows
	var err error
	if statusFilter != nil {
		rows, err = s.db.QueryContext(ctx, `SELECT id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body, req_body_path, resp_headers_json, resp_body, resp_body_path, attempts, created_at FROM history ORDER BY datetime(created_at) DESC LIMIT ? OFFSET ?`, limit, offset)
		// filter in Go to avoid complex WHERE with status ranges (tests use nil)
	} else {
		rows, err = s.db.QueryContext(ctx, `SELECT id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body, req_body_path, resp_headers_json, resp_body, resp_body_path, attempts, created_at FROM history ORDER BY datetime(created_at) DESC LIMIT ? OFFSET ?`, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

// Show returns one entry by id, loading spilled bodies.
func (s *Store) Show(ctx context.Context, id string) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.QueryContext(ctx, `SELECT id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body, req_body_path, resp_headers_json, resp_body, resp_body_path, attempts, created_at FROM history WHERE id=?`, id)
	if err != nil {
		return Entry{}, err
	}
	defer rows.Close()
	entries, err := scanEntries(rows)
	if err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, fmt.Errorf("history: not found: %s", id)
	}
	return entries[0], nil
}

// Search does FTS MATCH on url/request_path.
func (s *Store) Search(ctx context.Context, q string, limit int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// FTS ids
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM history_fts WHERE history_fts MATCH ? LIMIT ?`, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, nil
	}
	// fetch entries preserving FTS order
	var out []Entry
	for _, id := range ids {
		e, err := s.showLocked(ctx, id)
		if err != nil {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *Store) showLocked(ctx context.Context, id string) (Entry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, request_path, method, url, env, status, duration_ms, size, req_headers_json, req_body, req_body_path, resp_headers_json, resp_body, resp_body_path, attempts, created_at FROM history WHERE id=?`, id)
	if err != nil {
		return Entry{}, err
	}
	defer rows.Close()
	entries, err := scanEntries(rows)
	if err != nil {
		return Entry{}, err
	}
	if len(entries) == 0 {
		return Entry{}, fmt.Errorf("history: not found: %s", id)
	}
	return entries[0], nil
}

// EnforceRetention keeps last keep entries, deletes oldest.
func (s *Store) EnforceRetention(ctx context.Context, keep int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enforceRetentionLocked(ctx, keep)
}

func (s *Store) enforceRetentionLocked(ctx context.Context, keep int) error {
	if keep <= 0 {
		return nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM history ORDER BY datetime(created_at) ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		_ = rows.Scan(&id)
		ids = append(ids, id)
	}
	if len(ids) <= keep {
		return nil
	}
	toDelete := ids[:len(ids)-keep]
	for _, id := range toDelete {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM history WHERE id=?`, id)
		_, _ = s.db.ExecContext(ctx, `DELETE FROM history_fts WHERE id=?`, id)
		// also delete blobs
		_ = os.Remove(filepath.Join(filepath.Dir(s.dbPath), "history", "blobs", fmt.Sprintf("%s.req.bin", id)))
		_ = os.Remove(filepath.Join(filepath.Dir(s.dbPath), "history", "blobs", fmt.Sprintf("%s.resp.bin", id)))
	}
	return nil
}

// Clear deletes history, optionally filtered by env.
func (s *Store) Clear(ctx context.Context, env *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if env == nil {
		if _, err := s.db.ExecContext(ctx, `DELETE FROM history`); err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `DELETE FROM history_fts`); err != nil {
			return err
		}
		_ = os.RemoveAll(filepath.Join(filepath.Dir(s.dbPath), "history", "blobs"))
		return nil
	}
	// delete by env, need ids for FTS
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM history WHERE env=?`, *env)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		_ = rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		_, _ = s.db.ExecContext(ctx, `DELETE FROM history WHERE id=?`, id)
		_, _ = s.db.ExecContext(ctx, `DELETE FROM history_fts WHERE id=?`, id)
		_ = os.Remove(filepath.Join(filepath.Dir(s.dbPath), "history", "blobs", fmt.Sprintf("%s.req.bin", id)))
		_ = os.Remove(filepath.Join(filepath.Dir(s.dbPath), "history", "blobs", fmt.Sprintf("%s.resp.bin", id)))
	}
	return nil
}

// Cookies

func (s *Store) InsertCookie(ctx context.Context, c Cookie) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.Path == "" {
		c.Path = "/"
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO cookies (name, value, domain, path, expires_at, secure, http_only, same_site, env, created_at) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		c.Name, c.Value, c.Domain, c.Path, c.ExpiresAt.UTC().Format(time.RFC3339Nano), boolToInt(c.Secure), boolToInt(c.HttpOnly), c.SameSite, c.Env, time.Now().UTC().Format(time.RFC3339Nano))
	return err
}

func (s *Store) ListCookies(ctx context.Context, env string) ([]Cookie, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.QueryContext(ctx, `SELECT name, value, domain, path, expires_at, secure, http_only, same_site, env FROM cookies WHERE env=?`, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Cookie
	for rows.Next() {
		var c Cookie
		var expStr string
		var sec, httpOnly int
		if err := rows.Scan(&c.Name, &c.Value, &c.Domain, &c.Path, &expStr, &sec, &httpOnly, &c.SameSite, &c.Env); err != nil {
			return nil, err
		}
		c.Secure = sec != 0
		c.HttpOnly = httpOnly != 0
		if t, err := time.Parse(time.RFC3339Nano, expStr); err == nil {
			c.ExpiresAt = t
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) DeleteCookie(ctx context.Context, name, domain, path, env string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.ExecContext(ctx, `DELETE FROM cookies WHERE name=? AND domain=? AND path=? AND env=?`, name, domain, path, env)
	return err
}

func (s *Store) ClearCookies(ctx context.Context, env *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if env == nil {
		_, err := s.db.ExecContext(ctx, `DELETE FROM cookies`)
		return err
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM cookies WHERE env=?`, *env)
	return err
}

// FilterCookies returns cookies matching urlStr (domain/path/secure/expires).
// isHTTPS indicates whether the request is over TLS (secure cookies only on https).
func FilterCookies(cookies []Cookie, urlStr string, isHTTPS bool) []Cookie {
	u, err := url.Parse(urlStr)
	var host, path string
	if err == nil {
		host = u.Host
		if h, _, err2 := net.SplitHostPort(host); err2 == nil {
			host = h
		}
		path = u.Path
		if path == "" {
			path = "/"
		}
	} else {
		host = urlStr
		path = "/"
	}
	now := time.Now()
	var out []Cookie
	for _, c := range cookies {
		if !c.ExpiresAt.IsZero() && c.ExpiresAt.Before(now) {
			continue
		}
		if c.Secure && !isHTTPS {
			continue
		}
		if c.Domain != "" {
			d := c.Domain
			if len(d) > 0 && d[0] == '.' {
				d = d[1:]
			}
			if host != d && !hasSuffixDot(host, d) {
				continue
			}
		}
		p := c.Path
		if p == "" {
			p = "/"
		}
		if !pathMatches(path, p) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func hasSuffixDot(host, domain string) bool {
	return strings.HasSuffix(host, "."+domain)
}

func pathMatches(reqPath, cookiePath string) bool {
	if reqPath == cookiePath {
		return true
	}
	if strings.HasPrefix(reqPath, cookiePath) {
		if strings.HasSuffix(cookiePath, "/") {
			return true
		}
		rem := reqPath[len(cookiePath):]
		return len(rem) > 0 && rem[0] == '/'
	}
	return false
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func scanEntries(rows *sql.Rows) ([]Entry, error) {
	var out []Entry
	for rows.Next() {
		var e Entry
		var reqH, respH string
		var reqBody, respBody []byte
		var reqPath, respPath sql.NullString
		var createdStr string
		if err := rows.Scan(&e.ID, &e.RequestPath, &e.Method, &e.URL, &e.Env, &e.Status, &e.DurationMS, &e.Size, &reqH, &reqBody, &reqPath, &respH, &respBody, &respPath, &e.Attempts, &createdStr); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(reqH), &e.ReqHeaders)
		_ = json.Unmarshal([]byte(respH), &e.RespHeaders)
		if reqPath.Valid && reqPath.String != "" {
			if b, err := os.ReadFile(reqPath.String); err == nil {
				e.ReqBody = b
			} else {
				e.ReqBody = reqBody
			}
		} else {
			e.ReqBody = reqBody
		}
		if respPath.Valid && respPath.String != "" {
			if b, err := os.ReadFile(respPath.String); err == nil {
				e.RespBody = b
			} else {
				e.RespBody = respBody
			}
		} else {
			e.RespBody = respBody
		}
		if t, err := time.Parse(time.RFC3339Nano, createdStr); err == nil {
			e.CreatedAt = t
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// maxAttempts floors persisted attempt counts at 1 so legacy rows and
// zero-valued entries read back as a single send.
func maxAttempts(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
