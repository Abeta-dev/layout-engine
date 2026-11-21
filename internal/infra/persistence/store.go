// Package persistence implements the Clean Architecture domain.Store adapters
// for embedded pure-Go SQLite (default) and production PostgreSQL.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/abeta-dev/layout-engine/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

// Dialect selects the SQL database engine and placeholder/type conventions.
type Dialect string

const (
	DialectSQLite   Dialect = "sqlite"
	DialectPostgres Dialect = "postgres"
)

// utcTimeLayout is fixed-width RFC3339Nano in UTC so SQLite TEXT ordering
// matches chronological order down to nanosecond precision.
const utcTimeLayout = "2006-01-02T15:04:05.000000000Z"

// Config configures a SQLStore connection and migration behavior.
type Config struct {
	Dialect     Dialect
	DSN         string
	AutoMigrate bool
}

// SQLStore implements domain.Store across both SQLite and PostgreSQL.
type SQLStore struct {
	db      *sql.DB
	dialect Dialect
}

var _ domain.Store = (*SQLStore)(nil)

// Open initializes a SQLStore for SQLite (default) or PostgreSQL, configures
// connection pool settings, verifies connectivity, and optionally runs migrations.
func Open(ctx context.Context, cfg Config) (*SQLStore, error) {
	dialect := Dialect(strings.ToLower(strings.TrimSpace(string(cfg.Dialect))))
	if dialect == "" {
		dialect = DialectSQLite
	}

	var (
		driverName string
		dsn        = strings.TrimSpace(cfg.DSN)
	)

	switch dialect {
	case DialectSQLite:
		driverName = "sqlite"
		if dsn == "" {
			dsn = ":memory:"
		}
	case DialectPostgres:
		driverName = "pgx"
		if dsn == "" {
			return nil, errors.New("persistence: postgres DSN is required")
		}
	default:
		return nil, fmt.Errorf("persistence: unsupported dialect %q", cfg.Dialect)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("persistence: open %s: %w", dialect, err)
	}

	switch dialect {
	case DialectSQLite:
		// Single open connection eliminates SQLITE_BUSY lock contention under -race.
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)
	case DialectPostgres:
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(30 * time.Minute)
	}

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("persistence: ping %s: %w", dialect, err)
	}

	if dialect == DialectSQLite {
		if err := configureSQLitePragmas(ctx, db, dsn); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	store := &SQLStore{
		db:      db,
		dialect: dialect,
	}

	if cfg.AutoMigrate {
		if err := store.Migrate(ctx); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("persistence: auto-migrate: %w", err)
		}
	}

	return store, nil
}

func configureSQLitePragmas(ctx context.Context, db *sql.DB, dsn string) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}
	if !isSQLiteMemoryDSN(dsn) {
		pragmas = append(pragmas,
			"PRAGMA journal_mode = WAL;",
			"PRAGMA synchronous = NORMAL;",
		)
	}
	for _, stmt := range pragmas {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("persistence: sqlite pragma %q: %w", stmt, err)
		}
	}
	return nil
}

func isSQLiteMemoryDSN(dsn string) bool {
	lower := strings.ToLower(strings.TrimSpace(dsn))
	return lower == "" ||
		lower == ":memory:" ||
		strings.HasPrefix(lower, "file::memory:") ||
		strings.Contains(lower, "mode=memory")
}

// Dialect returns the active SQL dialect of the store.
func (s *SQLStore) Dialect() Dialect {
	return s.dialect
}

// Ping verifies database connectivity.
func (s *SQLStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// Close closes the underlying SQL connection pool.
func (s *SQLStore) Close() error {
	return s.db.Close()
}

// rebind converts SQLite-style `?` bind placeholders into PostgreSQL `$1, $2, ...`
// placeholders when targeting PostgreSQL. Single-quoted literals are preserved untouched.
func (s *SQLStore) rebind(query string) string {
	return rebindQuery(s.dialect, query)
}

func rebindQuery(dialect Dialect, query string) string {
	if dialect != DialectPostgres {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 16)

	argIdx := 1
	inSingleQuote := false
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' {
			b.WriteByte(ch)
			if inSingleQuote {
				// Escaped single quote ('') inside a SQL literal stays in literal.
				if i+1 < len(query) && query[i+1] == '\'' {
					i++
					b.WriteByte('\'')
				} else {
					inSingleQuote = false
				}
			} else {
				inSingleQuote = true
			}
			continue
		}
		if ch == '?' && !inSingleQuote {
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(argIdx))
			argIdx++
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func (s *SQLStore) formatTime(t time.Time) any {
	if t.IsZero() {
		t = time.Now().UTC()
	} else {
		t = t.UTC()
	}
	if s.dialect == DialectPostgres {
		return t
	}
	return t.Format(utcTimeLayout)
}

func parseTime(raw any) (time.Time, error) {
	switch v := raw.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v.UTC(), nil
	case []byte:
		return parseTimeString(string(v))
	case string:
		return parseTimeString(v)
	default:
		return time.Time{}, fmt.Errorf("persistence: unsupported time type %T", raw)
	}
}

func parseTimeString(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("persistence: parse timestamp %q", raw)
}

func marshalJSON(v any, emptyFallback string) (string, error) {
	if isNilValue(v) {
		return emptyFallback, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if string(raw) == "null" {
		return emptyFallback, nil
	}
	return string(raw), nil
}

func marshalNullableJSON(v any) (sql.NullString, error) {
	if isNilValue(v) {
		return sql.NullString{Valid: false}, nil
	}
	rv := reflect.ValueOf(v)
	if (rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice) && rv.Len() == 0 {
		return sql.NullString{Valid: false}, nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return sql.NullString{}, err
	}
	if string(raw) == "null" {
		return sql.NullString{Valid: false}, nil
	}
	return sql.NullString{String: string(raw), Valid: true}, nil
}

func unmarshalJSON(raw string, dst any) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "null" {
		return nil
	}
	return json.Unmarshal([]byte(trimmed), dst)
}

func unmarshalNullableJSON(ns sql.NullString, dst any) error {
	if !ns.Valid {
		return nil
	}
	return unmarshalJSON(ns.String, dst)
}

func isNilValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key value violates unique constraint") ||
		strings.Contains(msg, "SQLSTATE 23505")
}

type rowScanner interface {
	Scan(dest ...any) error
}

func collectRows[T any](rows *sql.Rows, scanFn func(rowScanner) (T, error)) ([]T, error) {
	defer func() { _ = rows.Close() }()
	out := make([]T, 0)
	for rows.Next() {
		item, err := scanFn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *SQLStore) execDelete(ctx context.Context, label, query string, args ...any) error {
	res, err := s.db.ExecContext(ctx, s.rebind(query), args...)
	if err != nil {
		return fmt.Errorf("persistence: delete %s: %w", label, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("persistence: delete %s rows affected: %w", label, err)
	}
	if affected == 0 {
		return fmt.Errorf("persistence: %s: %w", label, domain.ErrNotFound)
	}
	return nil
}

func parseTimestamps(rawCreated, rawUpdated any) (time.Time, time.Time, error) {
	created, err := parseTime(rawCreated)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	updated, err := parseTime(rawUpdated)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return created, updated, nil
}
