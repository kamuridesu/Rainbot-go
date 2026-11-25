package sticker

import (
	"bytes"
	"encoding/binary"
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
	data, err := GenerateMetadata(`Rain"bot`, "MyPack")
	if err != nil {
		t.Fatalf("GenerateMetadata() error = %v", err)
	}

	// Walk the TIFF structure the way a strict EXIF reader would.
	be := binary.BigEndian
	if len(data) < 26 || string(data[:2]) != "MM" || be.Uint16(data[2:4]) != 42 {
		t.Fatalf("bad TIFF header: % x", data[:min(8, len(data))])
	}
	ifd := int(be.Uint32(data[4:8]))
	if n := be.Uint16(data[ifd:]); n != 1 {
		t.Fatalf("entries = %d, want 1", n)
	}
	entry := data[ifd+2 : ifd+14]
	tag, typ, count, offset := be.Uint16(entry[0:2]), be.Uint16(entry[2:4]), be.Uint32(entry[4:8]), be.Uint32(entry[8:12])
	if tag != 0x5741 || typ != 7 {
		t.Errorf("tag=%#x type=%d, want 0x5741/7", tag, typ)
	}
	if next := be.Uint32(data[ifd+14 : ifd+18]); next != 0 {
		t.Errorf("next IFD pointer = %#x, want 0 (JSON must not overlap it)", next)
	}
	if int(offset) < ifd+18 || int(offset)+int(count) != len(data) {
		t.Fatalf("value at %d+%d does not fit after the IFD in %d bytes", offset, count, len(data))
	}

	var parsed struct {
		Name      string `json:"sticker-pack-name"`
		Publisher string `json:"sticker-pack-publisher"`
	}
	if err := json.Unmarshal(data[offset:offset+count], &parsed); err != nil {
		t.Fatalf("failed to parse embedded JSON metadata: %v", err)
	}
	if parsed.Name != `Rain"bot` || parsed.Publisher != "MyPack" {
		t.Errorf("parsed metadata = %+v", parsed)
	}
}

// TestGenerateMetadataMatchesOfficialLayout compares the header against the
// one in a sticker made by WhatsApp's own sticker maker (JSON length 0x77).
func TestGenerateMetadataMatchesOfficialLayout(t *testing.T) {
	official := []byte{0x4d, 0x4d, 0x00, 0x2a, 0x00, 0x00, 0x00, 0x08, 0x00, 0x01, 0x57, 0x41, 0x00, 0x07, 0x00, 0x00, 0x00, 0x77, 0x00, 0x00, 0x00, 0x1a, 0x00, 0x00, 0x00, 0x00}

	data, err := GenerateMetadata("a", "b")
	if err != nil {
		t.Fatal(err)
	}
	got := append([]byte{}, data[:26]...)
	copy(got[14:18], official[14:18]) // the JSON length differs
	if !bytes.Equal(got, official) {
		t.Errorf("header = % x\nwant     % x", got, official)
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
