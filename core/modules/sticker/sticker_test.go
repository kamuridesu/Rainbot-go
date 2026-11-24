package sticker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"testing"
)

func TestGenerateMetadata(t *testing.T) {
	data, err := GenerateMetadata("Rainbot", "MyPack")
	if err != nil {
		t.Fatalf("GenerateMetadata() error = %v", err)
	}

	const headerLen = 14 + 2 + 6
	if len(data) <= headerLen {
		t.Fatalf("GenerateMetadata() output too short: %d bytes", len(data))
	}

	jsonPart := data[headerLen:]
	var parsed struct {
		Name      string `json:"sticker-pack-name"`
		Publisher string `json:"sticker-pack-publisher"`
	}
	if err := json.Unmarshal(jsonPart, &parsed); err != nil {
		t.Fatalf("failed to parse embedded JSON metadata: %v (raw: %q)", err, jsonPart)
	}
	if parsed.Name != "Rainbot" || parsed.Publisher != "MyPack" {
		t.Errorf("parsed metadata = %+v, want Name=Rainbot Publisher=MyPack", parsed)
	}

	jsonLen := int(data[14]) | int(data[15])<<8
	if jsonLen != len(jsonPart) {
		t.Errorf("encoded length = %d, want %d", jsonLen, len(jsonPart))
	}
}

func TestCreateTempFileAndDeleteTmpFile(t *testing.T) {
	content := []byte("hello sticker")

	path, err := CreateTempFile(content)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read created temp file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("temp file content = %q, want %q", got, content)
	}

	if err := DeleteTmpFile(path); err != nil {
		t.Fatalf("DeleteTmpFile() error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected temp file to be gone after DeleteTmpFile()")
	}
}

func TestNew(t *testing.T) {
	s := New("author", "pack", []byte("data"), StickerSquash)
	if s.Author != "author" || s.Pack != "pack" || string(s.Data) != "data" || s.Type != StickerSquash {
		t.Errorf("New() = %+v, want fields to match constructor args", s)
	}
}

func encodeTestPNGForSticker(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	return buf.Bytes()
}

func requireExternalTools(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping sticker conversion integration test")
	}
	if _, err := exec.LookPath("webpmux"); err != nil {
		t.Skip("webpmux not available, skipping sticker conversion integration test")
	}
}

func TestStickerConvert(t *testing.T) {
	requireExternalTools(t)

	png := encodeTestPNGForSticker(t)

	tests := []StickerType{StickerOriginal, StickerSquash, StickerTransparent}
	for _, st := range tests {
		s := New("Rainbot", "TestPack", png, st)
		out, err := s.Convert()
		if err != nil {
			t.Fatalf("Convert() type=%v error = %v", st, err)
		}
		if len(out) < 12 || string(out[0:4]) != "RIFF" || string(out[8:12]) != "WEBP" {
			t.Errorf("Convert() type=%v output missing RIFF/WEBP header", st)
		}
	}
}

func TestStickerConvertHighMotionFitsLimit(t *testing.T) {
	requireExternalTools(t)
	if testing.Short() {
		t.Skip("slow: encodes several times")
	}

	in := t.TempDir() + "/in.mp4"
	if b, err := exec.Command("ffmpeg", "-v", "error",
		"-f", "lavfi", "-i", "mandelbrot=size=720x720:rate=30", "-t", "6",
		"-c:v", "libx264", "-b:v", "600k", "-pix_fmt", "yuv420p", in).CombinedOutput(); err != nil {
		t.Fatalf("generate video: %v: %s", err, b)
	}
	video, err := os.ReadFile(in)
	if err != nil {
		t.Fatal(err)
	}

	for _, st := range []StickerType{StickerSquash, StickerOriginal} {
		out, err := New("Rainbot", "TestPack", video, st).Convert()
		if err != nil {
			t.Fatalf("Convert() type=%v error = %v", st, err)
		}
		if len(out) > MaxStickerSize {
			t.Errorf("type=%v size = %d, want <= %d", st, len(out), MaxStickerSize)
		}

		f := t.TempDir() + "/out.webp"
		if err := os.WriteFile(f, out, 0o644); err != nil {
			t.Fatal(err)
		}
		info, err := exec.Command("webpmux", "-info", f).CombinedOutput()
		if err != nil {
			t.Fatalf("webpmux -info: %v: %s", err, info)
		}
		var frames, minFrames int
		for _, line := range bytes.Split(info, []byte("\n")) {
			if n, _ := fmt.Sscanf(string(line), "Number of frames: %d", &frames); n == 1 {
				break
			}
		}
		minFrames = 5 * encodeLadder[len(encodeLadder)-1].FPS
		if frames < minFrames {
			t.Errorf("type=%v frames = %d, want >= %d (animation truncated)", st, frames, minFrames)
		}
	}
}

func TestNextAttempt(t *testing.T) {
	last := len(encodeLadder) - 1
	tests := []struct {
		name    string
		i, size int
		want    int
	}{
		{"slightly over goes one step", 0, MaxStickerSize + 1, 1},
		{"1.8x skips one step", 0, MaxStickerSize * 18 / 10, 2},
		{"4x skips two steps", 0, MaxStickerSize * 4, 3},
		{"skipping never jumps past last step", last - 2, MaxStickerSize * 4, last},
	}
	if got := nextAttempt(last, MaxStickerSize*4); got <= last {
		t.Errorf("after last step: nextAttempt = %d, want > %d to end the loop", got, last)
	}
	for _, tt := range tests {
		if got := nextAttempt(tt.i, tt.size); got != tt.want {
			t.Errorf("%s: nextAttempt(%d, %d) = %d, want %d", tt.name, tt.i, tt.size, got, tt.want)
		}
	}
}
