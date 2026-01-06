package storage

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var (
	ErrNotExists = errors.New("File does not exists")
	ErrReadOnly  = errors.New("File in read only mode")
)

type FileStorage interface {
	Exists(ctx context.Context, filename string) (bool, error)
	Delete(ctx context.Context, filename string) error
	Read(ctx context.Context, filename string) ([]byte, error)
	Write(ctx context.Context, filename string, data []byte) error
}

type OpenFileMode string

const (
	ModeWrite    OpenFileMode = "write"
	ModeReadOnly OpenFileMode = "readonly"
	ModeAppend   OpenFileMode = "append"
)

func NewStorage() FileStorage {
	storageType := os.Getenv("STORAGE_TYPE")
	switch storageType {
	case "LOCAL":
		return NewLocalFileStorage()
	case "S3":
		return NewS3FileStorage()
	default:
		return NewLocalFileStorage()
	}
}

type File struct {
	Name    string
	mode    OpenFileMode
	storage FileStorage
}

func RandomFilename(ext string) string {
	tmp := make([]rune, 20)
	for i := range 20 {
		tmp[i] = letters[rand.IntN(len(letters))]
	}
	return string(tmp) + "." + ext
}

func NewFile(filename string, mode ...OpenFileMode) *File {
	m := ModeWrite
	if len(mode) > 0 {
		m = mode[0]
	}

	return &File{
		Name:    filename,
		mode:    m,
		storage: NewStorage(),
	}
}

func (f *File) Read(ctx context.Context) ([]byte, error) {
	_, err := f.storage.Exists(ctx, f.Name)
	if err != nil {
		return nil, err
	}

	return f.storage.Read(ctx, f.Name)
}

func (f *File) Delete(ctx context.Context) error {
	return f.storage.Delete(ctx, f.Name)
}

func (f *File) Write(ctx context.Context, data []byte) error {
	if f.mode == ModeReadOnly {
		return ErrReadOnly
	}
	if f.mode == ModeAppend {
		content, err := f.storage.Read(ctx, f.Name)
		if err != nil {
			return err
		}
		data = append(content, data...)
	}
	return f.storage.Write(ctx, f.Name, data)
}

func (f *File) WriteString(ctx context.Context, data string) error {
	return f.Write(ctx, []byte(data))
}

func (f *File) Exists(ctx context.Context) (bool, error) {
	return f.storage.Exists(ctx, f.Name)
}
