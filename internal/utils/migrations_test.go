package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadScriptsOrdersNumericallyByPrefix(t *testing.T) {
	dir := t.TempDir()

	names := []string{
		"10_create_token.sql",
		"00.sql",
		"2_create_quotly.sql",
		"1_create_messages.sql",
		"not_a_migration.txt",
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("SELECT 1;"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := LoadScripts(dir)
	if err != nil {
		t.Fatalf("LoadScripts() error = %v", err)
	}

	var got []string
	for _, e := range entries {
		got = append(got, e.Name())
	}

	want := []string{
		"00.sql",
		"1_create_messages.sql",
		"2_create_quotly.sql",
		"10_create_token.sql",
	}

	if len(got) != len(want) {
		t.Fatalf("LoadScripts() returned %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("LoadScripts()[%d] = %q, want %q (full result: %v)", i, got[i], want[i], got)
		}
	}
}

func TestLoadScriptsMissingDirectory(t *testing.T) {
	_, err := LoadScripts(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected an error for a missing directory, got nil")
	}
}
