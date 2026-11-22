package migrations_test

import (
	"io/fs"
	"testing"

	"github.com/abeta-dev/layout-engine/migrations"
)

func TestEmbeddedSQLMigrationsExist(t *testing.T) {
	t.Parallel()

	for _, path := range []string{
		"sqlite/0001_initial_schema.sql",
		"postgres/0001_initial_schema.sql",
	} {
		data, err := fs.ReadFile(migrations.Files, path)
		if err != nil {
			t.Fatalf("read embedded migration %s: %v", path, err)
		}
		if len(data) == 0 {
			t.Fatalf("embedded migration %s must not be empty", path)
		}
	}
}
