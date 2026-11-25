package sticker

import (
	"cmp"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/kamuridesu/rainbot-go/internal/storage"
)

const MaxStickerSize = 500 * 1024

var ErrStickerTooLarge = errors.New("sticker exceeds size limit even at lowest quality")

type encodeAttempt struct {
	FPS     int
	Quality int
}

var encodeLadder = []encodeAttempt{
	{FPS: 20, Quality: 75},
	{FPS: 15, Quality: 50},
	{FPS: 12, Quality: 35},
	{FPS: 10, Quality: 20},
	{FPS: 10, Quality: 10},
	{FPS: 8, Quality: 5},
}

const (
	scaleSquash      = "scale=512:512"
	scaleTransparent = "scale=512:512:force_original_aspect_ratio=decrease,pad=512:512:-1:-1:color=0x00000000"
)

// nextAttempt picks the next ladder step after one that produced size bytes.
// Each step shrinks output by roughly 1.3-1.6x, so far-oversized results skip
// steps that could not fit anyway; this saves whole encodes on long clips.
func nextAttempt(i, size int) int {
	next := i + 1
	switch ratio := float64(size) / MaxStickerSize; {
	case ratio > 3:
		next += 2
	case ratio > 1.6:
		next++
	}
	// Never skip past the last step: it is the only one that may still fit.
	if next >= len(encodeLadder) && i < len(encodeLadder)-1 {
		next = len(encodeLadder) - 1
	}
	return next
}

func encodeWebp(input, scale string, length time.Duration, a encodeAttempt) ([]byte, error) {
	output := fmt.Sprintf("%s-%d-%d.webp", input, a.FPS, a.Quality)
	defer DeleteTmpFile(output)

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", input,
		"-t", strconv.FormatFloat(length.Seconds(), 'f', 3, 64),
		"-an",
		"-vcodec", "libwebp",
		"-q:v", strconv.Itoa(a.Quality),
		"-compression_level", "4",
		"-vf", fmt.Sprintf("%s,fps=%d", scale, a.FPS),
		"-loop", "0",
		"-f", "webp",
		output,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("FFmpeg conversion failed", "error", err.Error(), "output", string(out))
		return nil, fmt.Errorf("ffmpeg error: %w", err)
	}

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

// stickerExifTag is the private EXIF tag WhatsApp reads sticker metadata from.
const stickerExifTag = 0x5741

// GenerateMetadata builds the EXIF block WhatsApp reads sticker pack info from:
// a big-endian TIFF header with a single IFD holding one UNDEFINED entry (tag
// 0x5741) that points at the JSON. The layout matches stickers made by official
// clients. The previous header omitted the next-IFD pointer, so the JSON
// overlapped it; regular stickers tolerated that, but clients refused to play
// the sound of audio stickers carrying it.
func GenerateMetadata(author, pack string) ([]byte, error) {
	metadata, err := json.Marshal(struct {
		Name      string `json:"sticker-pack-name"`
		Publisher string `json:"sticker-pack-publisher"`
	}{author, pack})
	if err != nil {
		return nil, err
	}

	const dataOffset = 8 + 2 + 12 + 4 // TIFF header, entry count, one entry, next-IFD pointer

	be := binary.BigEndian
	buf := make([]byte, 0, dataOffset+len(metadata))
	buf = append(buf, 'M', 'M')
	buf = be.AppendUint16(buf, 42) // TIFF magic
	buf = be.AppendUint32(buf, 8)  // offset of the first IFD
	buf = be.AppendUint16(buf, 1)  // number of entries
	buf = be.AppendUint16(buf, stickerExifTag)
	buf = be.AppendUint16(buf, 7) // type UNDEFINED
	buf = be.AppendUint32(buf, uint32(len(metadata)))
	buf = be.AppendUint32(buf, dataOffset)
	buf = be.AppendUint32(buf, 0) // no next IFD
	return append(buf, metadata...), nil
}

type StickerType int

const (
	StickerOriginal StickerType = iota
	StickerSquash
	StickerTransparent
)

const DefaultStickerLength = 10 * time.Second

type Sticker struct {
	Author string
	Pack   string
	Data   []byte
	Type   StickerType
	Length time.Duration
}

func New(author, pack string, data []byte, sType StickerType) *Sticker {
	return &Sticker{Author: author, Pack: pack, Data: data, Type: sType, Length: DefaultStickerLength}
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
	file, err := CreateTempFile(s.Data)
	if err != nil {
		return nil, err
	}
	defer DeleteTmpFile(file)

	scale := scaleTransparent
	if s.Type == StickerSquash {
		scale = scaleSquash
	}

	var output []byte
	length := cmp.Or(s.Length, DefaultStickerLength)
	for i := 0; i < len(encodeLadder); i = nextAttempt(i, len(output)) {
		attempt := encodeLadder[i]
		encoded, err := encodeWebp(file, scale, length, attempt)
		if err != nil {
			return nil, err
		}
		if output, err = s.addMetadata(encoded); err != nil {
			return nil, err
		}
		if len(output) <= MaxStickerSize {
			if i > 0 {
				slog.Info("Sticker reduced to fit size limit", "fps", attempt.FPS, "quality", attempt.Quality, "size", len(output))
			}
			return output, nil
		}
		slog.Info("Sticker too large, retrying with lower quality", "fps", attempt.FPS, "quality", attempt.Quality, "size", len(output))
	}

	return nil, fmt.Errorf("%w: %d KB", ErrStickerTooLarge, len(output)/1024)
}
