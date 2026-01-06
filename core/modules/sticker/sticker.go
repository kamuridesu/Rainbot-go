package sticker

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/kamuridesu/rainbot-go/internal/storage"
)

func ExecFFMpreg(input string) ([]byte, error) {
	output := fmt.Sprintf("%s-new.webp", input)
	defer DeleteTmpFile(output)

	vf := `scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:-1:-1:color=white@0.0,fps=20`

	cmd := exec.Command("ffmpeg",
		"-i", input,
		"-t", "5",
		"-an",
		"-vcodec", "libwebp",
		"-vf", vf,
		"-loop", "0",
		"-fs", "950K",
		"-f", "webp",
		output,
	)

	slog.Info("Iniciando conversao")

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("FFmpeg conversion failed", "error", err.Error(), "output", string(out))
		return nil, fmt.Errorf("ffmpeg error: %w", err)
	}

	slog.Info("Conversao concluida")

	return os.ReadFile(output)
}

func execWebpmux(imageData, exifData []byte) ([]byte, error) {
	imgFile, err := CreateTempFile(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for image: %w", err)
	}
	defer DeleteTmpFile(imgFile)

	exifFile, err := CreateTempFile(exifData)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for exif: %w", err)
	}
	defer DeleteTmpFile(exifFile)

	outputFile := fmt.Sprintf("%s-meta.webp", imgFile)
	defer DeleteTmpFile(outputFile)

	cmd := exec.Command("webpmux", "-set", "exif", exifFile, imgFile, "-o", outputFile)

	if out, err := cmd.CombinedOutput(); err != nil {
		slog.Error("webpmux failed", "error", err, "output", string(out))
		return nil, err
	}

	return os.ReadFile(outputFile)
}

func CreateTempFile(data []byte) (string, error) {
	tmpPath := os.TempDir()
	filename := storage.RandomFilename(fmt.Sprintf("%d", os.Getpid()))
	fullpath := path.Join(tmpPath, filename)
	err := os.WriteFile(fullpath, data, 0644)
	if err != nil {
		return "", err
	}
	return fullpath, nil
}

func GenerateMetadata(author, pack string) ([]byte, error) {
	metadata := fmt.Sprintf(`{"sticker-pack-name":"%s","sticker-pack-publisher":"%s"}`, author, pack)

	littleEndianHeader := []byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x41, 0x57, 0x07, 0x00}
	exifHeader := []byte{0x00, 0x00, 0x16, 0x00, 0x00, 0x00}

	jsonLen := len(metadata)
	lowByte := byte(jsonLen & 0xFF)
	highByte := byte((jsonLen >> 8) & 0xFF)

	var finalBuffer bytes.Buffer
	finalBuffer.Write(littleEndianHeader)
	finalBuffer.WriteByte(lowByte)
	finalBuffer.WriteByte(highByte)
	finalBuffer.Write(exifHeader)
	finalBuffer.Write([]byte(metadata))

	return finalBuffer.Bytes(), nil
}

type Sticker struct {
	Author string
	Pack   string
	Data   []byte
}

func New(author, pack string, data []byte) *Sticker {
	return &Sticker{author, pack, data}
}

func DeleteTmpFile(filename string) error {
	return os.Remove(filename)
}

func (s *Sticker) addMetadata(content []byte) ([]byte, error) {
	metadata, err := GenerateMetadata(s.Author, s.Pack)
	if err != nil {
		return nil, err
	}
	return execWebpmux(content, metadata)
}

func (s *Sticker) Convert() ([]byte, error) {
	slog.Info("Starting convertion")
	file, err := CreateTempFile(s.Data)
	if err != nil {
		return nil, err
	}

	slog.Info("Starting ffmpeg")
	output, err := ExecFFMpreg(file)
	if err != nil {
		return nil, err
	}

	slog.Info("adding metadata")
	return s.addMetadata(output)
}
