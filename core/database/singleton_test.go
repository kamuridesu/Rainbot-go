package database

import (
	"path/filepath"
	"testing"
)

func TestDatabaseSingletonLifecycle(t *testing.T) {
	if _, err := GetDatabaseSingleton(); err == nil {
		t.Fatal("expected an error before InitDatabaseSingleton has ever been called")
	}

	dbfile := filepath.Join(t.TempDir(), "test.db")
	s1, err := InitDatabaseSingleton("sqlite3", "file:"+dbfile+"?_foreign_keys=on", false)
	if err != nil {
		t.Fatalf("InitDatabaseSingleton() error = %v", err)
	}
	if s1.Chat == nil || s1.Member == nil || s1.Filter == nil || s1.Message == nil || s1.Quotly == nil {
		t.Errorf("expected all services to be wired, got %+v", s1)
	}

	s2, err := InitDatabaseSingleton("sqlite3", "not a real dsn", false)
	if err != nil {
		t.Fatalf("InitDatabaseSingleton() (cached call) error = %v", err)
	}
	if s2 != s1 {
		t.Error("expected the cached singleton to be returned on a subsequent call")
	}

	got, err := GetDatabaseSingleton()
	if err != nil {
		t.Fatalf("GetDatabaseSingleton() error = %v", err)
	}
	if got != s1 {
		t.Error("expected GetDatabaseSingleton() to return the initialized singleton")
	}

	got.Close()
}
