package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/abeta-dev/layout-engine/migrations"
)

const createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL
);`

// Migrate applies all embedded SQL migrations for the store's dialect in sorted
// filename order inside individual transactions. Already-applied versions are skipped.
func (s *SQLStore) Migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, createMigrationsTableSQL); err != nil {
		return fmt.Errorf("persistence: ensure schema_migrations: %w", err)
	}

	dir := string(s.dialect)
	entries, err := fs.ReadDir(migrations.Files, dir)
	if err != nil {
		return fmt.Errorf("persistence: read embedded migrations %q: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		applied, err := s.isMigrationApplied(ctx, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		rawSQL, err := fs.ReadFile(migrations.Files, path.Join(dir, name))
		if err != nil {
			return fmt.Errorf("persistence: read migration %s: %w", name, err)
		}

		if err := s.applyMigration(ctx, name, string(rawSQL)); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLStore) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var existing string
	err := s.db.QueryRowContext(
		ctx,
		s.rebind(`SELECT version FROM schema_migrations WHERE version = ?`),
		version,
	).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("persistence: check migration %s: %w", version, err)
	}
	return true, nil
}

func (s *SQLStore) applyMigration(ctx context.Context, version, script string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("persistence: begin migration %s: %w", version, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, script); err != nil {
		return fmt.Errorf("persistence: execute migration %s: %w", version, err)
	}

	appliedAt := time.Now().UTC().Format(utcTimeLayout)
	if _, err := tx.ExecContext(
		ctx,
		s.rebind(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`),
		version,
		appliedAt,
	); err != nil {
		return fmt.Errorf("persistence: record migration %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("persistence: commit migration %s: %w", version, err)
	}
	return nil
}
