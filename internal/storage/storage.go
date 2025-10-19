package storage

import (
	"errors"
	"log/slog"
	"math/rand/v2"
	"os"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var (
	ErrNotExists = errors.New("File does not exists")
	ErrReadOnly  = errors.New("File in read only mode")
)

type FileStorage interface {
	Exists(filename string) (bool, error)
	Delete(filename string) error
	Read(filename string) ([]byte, error)
	Write(filename string, data []byte) error
}

type LocalFileStorage struct{}

func NewLocalFileStorage() FileStorage {
	return &LocalFileStorage{}
}

func (s *LocalFileStorage) Write(filename string, data []byte) error {
	slog.Info("writing file")
	return os.WriteFile(filename, data, 0644)
}

func (s *LocalFileStorage) Delete(filename string) error {
	return os.Remove(filename)
}

func (s *LocalFileStorage) Exists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, ErrNotExists
	}
	return false, err
}

func (s *LocalFileStorage) Read(filename string) ([]byte, error) {
	return os.ReadFile(filename)
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

func (f *File) Read() ([]byte, error) {
	_, err := f.storage.Exists(f.Name)
	if err != nil {
		return nil, err
	}

	return f.storage.Read(f.Name)
}

func (f *File) Delete() error {
	return f.storage.Delete(f.Name)
}

func (f *File) Write(data []byte) error {
	if f.mode == ModeReadOnly {
		return ErrReadOnly
	}
	if f.mode == ModeAppend {
		content, err := f.storage.Read(f.Name)
		if err != nil {
			return err
		}
		data = append(content, data...)
	}
	return f.storage.Write(f.Name, data)
}

func (f *File) WriteString(data string) error {
	return f.Write([]byte(data))
}

func (f *File) Exists() (bool, error) {
	return f.storage.Exists(f.Name)
}
