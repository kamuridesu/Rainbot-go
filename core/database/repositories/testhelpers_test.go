package repositories_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/providers"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

func newTestDB(t *testing.T) *providers.Database {
	t.Helper()

	dbfile := filepath.Join(t.TempDir(), "test.db")
	db, err := providers.InitDB("sqlite3", "file:"+dbfile+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	t.Chdir(repoRoot(t))

	if err := utils.MigrateSqlite(db); err != nil {
		t.Fatalf("MigrateSqlite() error = %v", err)
	}

	return db
}
