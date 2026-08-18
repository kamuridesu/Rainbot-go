package utils

import (
	"os"
	"testing"
)

func TestReadDotEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	content := "FOO=bar\nBAZ=has=equals=signs\n\nQUOTED=value with spaces\n"
	if err := os.WriteFile(".env", []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	for _, key := range []string{"FOO", "BAZ", "QUOTED"} {
		os.Unsetenv(key)
		t.Cleanup(func(k string) func() { return func() { os.Unsetenv(k) } }(key))
	}

	ReadDotEnv()

	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO = %q, want %q", got, "bar")
	}
	if got := os.Getenv("BAZ"); got != "has=equals=signs" {
		t.Errorf("BAZ = %q, want %q", got, "has=equals=signs")
	}
	if got := os.Getenv("QUOTED"); got != "value with spaces" {
		t.Errorf("QUOTED = %q, want %q", got, "value with spaces")
	}
}

func TestReadDotEnvMissingFileIsNoop(t *testing.T) {
	t.Chdir(t.TempDir())
	ReadDotEnv()
}
