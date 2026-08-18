package storage

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRandomFilename(t *testing.T) {
	name := RandomFilename("png")

	if !strings.HasSuffix(name, ".png") {
		t.Errorf("RandomFilename(%q) = %q, want suffix %q", "png", name, ".png")
	}

	base := strings.TrimSuffix(name, ".png")
	if len(base) != 20 {
		t.Errorf("RandomFilename() base length = %d, want 20 (got %q)", len(base), base)
	}
	for _, r := range base {
		if !strings.ContainsRune(string(letters), r) {
			t.Errorf("RandomFilename() contains unexpected character %q in %q", r, base)
		}
	}
}

func TestRandomFilenameIsNotConstant(t *testing.T) {
	seen := map[string]bool{}
	for range 50 {
		seen[RandomFilename("png")] = true
	}
	if len(seen) < 2 {
		t.Error("expected RandomFilename() to vary across calls, got the same value every time")
	}
}

func TestFileDefaultModeIsWrite(t *testing.T) {
	f := NewFile("somefile.png")
	if f.mode != ModeWrite {
		t.Errorf("default mode = %v, want %v", f.mode, ModeWrite)
	}
}

func TestFileWriteReadDeleteExistsRoundTrip(t *testing.T) {
	ctx := context.Background()
	t.Setenv("STORAGE_TYPE", "LOCAL")

	path := filepath.Join(t.TempDir(), "test-file.txt")
	f := NewFile(path)

	if exists, err := f.Exists(ctx); err == nil || exists {
		t.Errorf("Exists() before write = (%v, %v), want (false, ErrNotExists)", exists, err)
	}

	if err := f.Write(ctx, []byte("hello")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	exists, err := f.Exists(ctx)
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("expected file to exist after Write()")
	}

	got, err := f.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("Read() = %q, want %q", got, "hello")
	}

	if err := f.Delete(ctx); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if exists, _ := f.Exists(ctx); exists {
		t.Error("expected file to be gone after Delete()")
	}
}

func TestFileWriteReadOnlyModeRejected(t *testing.T) {
	ctx := context.Background()
	t.Setenv("STORAGE_TYPE", "LOCAL")

	path := filepath.Join(t.TempDir(), "test-file.txt")
	f := NewFile(path, ModeReadOnly)

	if err := f.Write(ctx, []byte("hello")); err != ErrReadOnly {
		t.Errorf("Write() in read-only mode error = %v, want %v", err, ErrReadOnly)
	}
}

func TestFileWriteAppendModeConcatenates(t *testing.T) {
	ctx := context.Background()
	t.Setenv("STORAGE_TYPE", "LOCAL")

	path := filepath.Join(t.TempDir(), "test-file.txt")

	writer := NewFile(path, ModeWrite)
	if err := writer.Write(ctx, []byte("hello ")); err != nil {
		t.Fatalf("initial Write() error = %v", err)
	}

	appender := NewFile(path, ModeAppend)
	if err := appender.Write(ctx, []byte("world")); err != nil {
		t.Fatalf("append Write() error = %v", err)
	}

	got, err := NewFile(path, ModeReadOnly).Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("Read() after append = %q, want %q", got, "hello world")
	}
}

func TestFileWriteStringWrapsWrite(t *testing.T) {
	ctx := context.Background()
	t.Setenv("STORAGE_TYPE", "LOCAL")

	path := filepath.Join(t.TempDir(), "test-file.txt")
	f := NewFile(path)

	if err := f.WriteString(ctx, "hello"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	got, err := f.Read(ctx)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(got) != "hello" {
		t.Errorf("Read() = %q, want %q", got, "hello")
	}
}

func TestFileReadNonexistentReturnsNotExistsWithoutTouchingStorage(t *testing.T) {
	ctx := context.Background()
	t.Setenv("STORAGE_TYPE", "LOCAL")

	path := filepath.Join(t.TempDir(), "does-not-exist.txt")
	f := NewFile(path, ModeReadOnly)

	if _, err := f.Read(ctx); err != ErrNotExists {
		t.Errorf("Read() of missing file error = %v, want %v", err, ErrNotExists)
	}
}
