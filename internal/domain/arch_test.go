package domain_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCleanArchitectureBoundaries enforces the Uncle Bob Dependency Rule:
// internal/domain and internal/engine must never import HTTP, Gin, SQL drivers,
// internal/infra, or internal/handler packages.
func TestCleanArchitectureBoundaries(t *testing.T) {
	t.Parallel()

	forbidden := []string{
		"github.com/gin-gonic/gin",
		"net/http",
		"database/sql",
		"modernc.org/sqlite",
		"github.com/jackc/pgx",
		"internal/infra",
		"internal/handler",
	}

	dirs := []string{
		".",
		filepath.Join("..", "engine"),
	}

	fset := token.NewFileSet()
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			t.Fatalf("read dir %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			fullPath := filepath.Join(dir, entry.Name())
			file, err := parser.ParseFile(fset, fullPath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("parse %s: %v", fullPath, err)
			}
			for _, imp := range file.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				for _, bad := range forbidden {
					if importPath == bad || strings.Contains(importPath, bad) {
						t.Errorf("Clean Architecture violation: %s imports forbidden package %q", fullPath, importPath)
					}
				}
			}
		}
	}
}
