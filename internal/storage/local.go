package storage

import (
	"context"
	"log/slog"
	"os"
)

type LocalFileStorage struct{}

func NewLocalFileStorage() FileStorage {
	return &LocalFileStorage{}
}

func (s *LocalFileStorage) Write(ctx context.Context, filename string, data []byte) error {
	slog.Info("writing file")
	return os.WriteFile(filename, data, 0644)
}

func (s *LocalFileStorage) Delete(ctx context.Context, filename string) error {
	return os.Remove(filename)
}

func (s *LocalFileStorage) Exists(ctx context.Context, filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, ErrNotExists
	}
	return false, err
}

func (s *LocalFileStorage) Read(ctx context.Context, filename string) ([]byte, error) {
	return os.ReadFile(filename)
}
