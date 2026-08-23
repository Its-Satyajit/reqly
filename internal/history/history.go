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
	"unicode"

	sqlcdb "github.com/Its-Satyajit/reqly/internal/history/db"

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

// Store is the SQLite store for history and cookies. SQL lives in db/*.sql
// and is compiled to typed Go by sqlc (db package); this struct owns the
// handle, locking, spill files, and row mapping only.
type Store struct {
	dbPath string
	db     *sql.DB
	q      *sqlcdb.Queries
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
	s := &Store{dbPath: dbPath, db: db, q: sqlcdb.New(db)}
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
	if err := s.q.InsertHistory(ctx, sqlcdb.InsertHistoryParams{
		ID:              e.ID,
		RequestPath:     sql.NullString{String: e.RequestPath, Valid: true},
		Method:          sql.NullString{String: e.Method, Valid: true},
		Url:             sql.NullString{String: e.URL, Valid: true},
		Env:             sql.NullString{String: e.Env, Valid: true},
		Status:          sql.NullInt64{Int64: int64(e.Status), Valid: true},
		DurationMs:      sql.NullInt64{Int64: int64(e.DurationMS), Valid: true},
		Size:            sql.NullInt64{Int64: int64(e.Size), Valid: true},
		ReqHeadersJson:  sql.NullString{String: string(reqH), Valid: true},
		ReqBody:         reqBodyBlob,
		ReqBodyPath:     sql.NullString{String: reqBodyPath, Valid: true},
		RespHeadersJson: sql.NullString{String: string(respH), Valid: true},
		RespBody:        respBodyBlob,
		RespBodyPath:    sql.NullString{String: respBodyPath, Valid: true},
		Attempts:        sql.NullInt64{Int64: int64(maxAttempts(e.Attempts)), Valid: true},
		CreatedAt:       e.CreatedAt.UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return err
	}
	// FTS
	if err := s.q.InsertHistoryFts(ctx, sqlcdb.InsertHistoryFtsParams{
		Url:         e.URL,
		RequestPath: e.RequestPath,
		ID:          e.ID,
	}); err != nil {
		return err
	}
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

// List returns entries ordered by created_at DESC.
func (s *Store) List(ctx context.Context, limit, offset int, statusFilter *int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.q.ListHistory(ctx, sqlcdb.ListHistoryParams{Limit: int64(limit), Offset: int64(offset)})
	if err != nil {
		return nil, err
	}
	out := make([]Entry, 0, len(rows))
	for _, row := range rows {
		entry, err := rowToEntry(row)
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}

// rowToEntry maps a generated history row onto the public Entry, hydrating
// spilled bodies from disk when the blob column is empty.
func rowToEntry(row sqlcdb.History) (Entry, error) {
	var e Entry
	e.ID = row.ID
	e.RequestPath = row.RequestPath.String
	e.Method = row.Method.String
	e.URL = row.Url.String
	e.Env = row.Env.String
	e.Status = int(row.Status.Int64)
	e.DurationMS = row.DurationMs.Int64
	e.Size = row.Size.Int64
	e.Attempts = int(row.Attempts.Int64)
	if t, err := time.Parse(time.RFC3339Nano, row.CreatedAt); err == nil {
		e.CreatedAt = t
	}
	_ = json.Unmarshal([]byte(row.ReqHeadersJson.String), &e.ReqHeaders)
	_ = json.Unmarshal([]byte(row.RespHeadersJson.String), &e.RespHeaders)
	e.ReqBody = loadSpillable(row.ReqBody, row.ReqBodyPath.String)
	e.RespBody = loadSpillable(row.RespBody, row.RespBodyPath.String)
	return e, nil
}

// loadSpillable prefers the on-disk spill file over the inline blob.
func loadSpillable(blob []byte, path string) []byte {
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			return b
		}
	}
	return blob
}

// Show returns one entry by id, loading spilled bodies.
func (s *Store) Show(ctx context.Context, id string) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.showLocked(ctx, id)
}

// ftsQuery converts free-text user input into a safe FTS5 MATCH expression.
// Every whitespace-separated token carrying at least one letter or digit
// becomes a double-quoted phrase (embedded quotes doubled) with prefix
// matching, so FTS5 operators, punctuation, and column-filter syntax typed by
// the user are treated as literal search text instead of query syntax.
func ftsQuery(q string) string {
	phrases := make([]string, 0, strings.Count(q, " ")+1)
	for _, tok := range strings.Fields(q) {
		if !strings.ContainsFunc(tok, unicode.IsLetter) && !strings.ContainsFunc(tok, unicode.IsDigit) {
			continue
		}
		phrases = append(phrases, `"`+strings.ReplaceAll(tok, `"`, `""`)+`"*`)
	}
	return strings.Join(phrases, " ")
}

// Search does FTS MATCH on url/request_path.
func (s *Store) Search(ctx context.Context, q string, limit int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	match := ftsQuery(q)
	if match == "" {
		return nil, nil
	}
	// FTS ids
	ids, err := s.q.SearchHistoryIds(ctx, sqlcdb.SearchHistoryIdsParams{
		Url:         match,
		RequestPath: match,
		Limit:       int64(limit),
	})
	if err != nil {
		return nil, err
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
	row, err := s.q.GetHistory(ctx, id)
	if err != nil {
		return Entry{}, fmt.Errorf("history: not found: %s", id)
	}
	return rowToEntry(row)
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
	ids, err := s.q.ListHistoryIdsAsc(ctx)
	if err != nil {
		return err
	}
	if len(ids) <= keep {
		return nil
	}
	toDelete := ids[:len(ids)-keep]
	for _, id := range toDelete {
		_ = s.q.DeleteHistoryById(ctx, id)
		_ = s.q.DeleteHistoryFtsById(ctx, id)
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
		if err := s.q.DeleteAllHistory(ctx); err != nil {
			return err
		}
		if err := s.q.DeleteAllHistoryFts(ctx); err != nil {
			return err
		}
		_ = os.RemoveAll(filepath.Join(filepath.Dir(s.dbPath), "history", "blobs"))
		return nil
	}
	// delete by env, need ids for FTS + blob cleanup
	ids, err := s.q.ListHistoryIdsByEnv(ctx, sql.NullString{String: *env, Valid: true})
	if err != nil {
		return err
	}
	for _, id := range ids {
		_ = s.q.DeleteHistoryById(ctx, id)
		_ = s.q.DeleteHistoryFtsById(ctx, id)
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
	return s.q.UpsertCookie(ctx, sqlcdb.UpsertCookieParams{
		Name:      nullStr(c.Name),
		Value:     nullStr(c.Value),
		Domain:    nullStr(c.Domain),
		Path:      nullStr(c.Path),
		ExpiresAt: c.ExpiresAt.UTC().Format(time.RFC3339Nano),
		Secure:    sql.NullInt64{Int64: int64(boolToInt(c.Secure)), Valid: true},
		HttpOnly:  sql.NullInt64{Int64: int64(boolToInt(c.HttpOnly)), Valid: true},
		SameSite:  nullStr(c.SameSite),
		Env:       nullStr(c.Env),
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func (s *Store) ListCookies(ctx context.Context, env string) ([]Cookie, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.q.ListCookiesByEnv(ctx, sql.NullString{String: env, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]Cookie, 0, len(rows))
	for _, row := range rows {
		c := Cookie{
			Name:     row.Name.String,
			Value:    row.Value.String,
			Domain:   row.Domain.String,
			Path:     row.Path.String,
			Secure:   row.Secure.Int64 != 0,
			HttpOnly: row.HttpOnly.Int64 != 0,
			SameSite: row.SameSite.String,
			Env:      row.Env.String,
		}
		if t, err := time.Parse(time.RFC3339Nano, row.ExpiresAt); err == nil {
			c.ExpiresAt = t
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) DeleteCookie(ctx context.Context, name, domain, path, env string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.q.DeleteCookie(ctx, sqlcdb.DeleteCookieParams{
		Name:   nullStr(name),
		Domain: nullStr(domain),
		Path:   nullStr(path),
		Env:    nullStr(env),
	})
}

func (s *Store) ClearCookies(ctx context.Context, env *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if env == nil {
		return s.q.DeleteAllCookies(ctx)
	}
	return s.q.DeleteCookiesByEnv(ctx, sql.NullString{String: *env, Valid: true})
}

// nullStr wraps a required string column value for generated Null params.
func nullStr(v string) sql.NullString {
	return sql.NullString{String: v, Valid: true}
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
