package storage

import (
	"database/sql"
	"errors"
	"time"

	_ "modernc.org/sqlite"

	"eedg-shortlink/models"
)

// Store — обёртка над SQLite.
type Store struct {
	db *sql.DB
}

// NewStore открывает/создаёт БД и применяет схему.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	schema := `
	CREATE TABLE IF NOT EXISTS links (
		code         TEXT PRIMARY KEY,
		long_url     TEXT NOT NULL,
		force_check  INTEGER NOT NULL DEFAULT 0,
		cache        BLOB,
		mime_type    TEXT,
		expires_at   DATETIME,
		created_at   DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_links_expires ON links(expires_at);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close закрывает соединение.
func (s *Store) Close() error { return s.db.Close() }

// InsertUnique пытается вставить запись, retry при коллизии кода.
// genCode вызывается на каждой попытке.
func (s *Store) InsertUnique(genCode func() (string, error), build func(code string) models.LinkRow, attempts int) (*models.LinkRow, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		code, err := genCode()
		if err != nil {
			return nil, err
		}
		row := build(code)
		err = s.insert(row)
		if err == nil {
			return &row, nil
		}
		if isUniqueViolation(err) {
			lastErr = err
			continue
		}
		return nil, err
	}
	return nil, lastErr
}

func (s *Store) insert(row models.LinkRow) error {
	_, err := s.db.Exec(
		`INSERT INTO links(code, long_url, force_check, cache, mime_type, expires_at, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		row.Code, row.LongURL, boolToInt(row.ForceCheck), row.Cache, row.MimeType,
		nullableTime(row.ExpiresAt), row.CreatedAt,
	)
	return err
}

// Get возвращает запись по коду или nil, если её нет.
func (s *Store) Get(code string) (*models.LinkRow, error) {
	row := s.db.QueryRow(
		`SELECT code, long_url, force_check, cache, mime_type, expires_at, created_at
		 FROM links WHERE code = ?`, code,
	)
	var (
		r       models.LinkRow
		forceIn int
		expires sql.NullTime
		mime    sql.NullString
	)
	err := row.Scan(&r.Code, &r.LongURL, &forceIn, &r.Cache, &mime, &expires, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.ForceCheck = forceIn == 1
	if expires.Valid {
		r.ExpiresAt = expires.Time
	}
	if mime.Valid {
		r.MimeType = mime.String
	}
	return &r, nil
}

// UpdateCache обновляет кеш записи.
func (s *Store) UpdateCache(code string, data []byte, mime string, expires time.Time) error {
	_, err := s.db.Exec(
		`UPDATE links SET cache = ?, mime_type = ?, expires_at = ? WHERE code = ?`,
		data, mime, expires, code,
	)
	return err
}

// ClearCache сбрасывает кеш записи.
func (s *Store) ClearCache(code string) error {
	return s.UpdateCache(code, nil, "", time.Time{})
}

// DeleteExpired очищает просроченный кеш, оставляя шортлинки.
func (s *Store) DeleteExpired() error {
	_, err := s.db.Exec(
		`UPDATE links SET cache = NULL, mime_type = NULL, expires_at = NULL
		 WHERE expires_at IS NOT NULL AND expires_at < ?`, time.Now().UTC(),
	)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func nullableTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite возвращает ошибку вида "UNIQUE constraint failed"
	return contains(err.Error(), "UNIQUE constraint failed")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
