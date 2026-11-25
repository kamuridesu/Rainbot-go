package sticker

import (
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func requireFFmpeg(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip(tool + " not available")
		}
	}
}

func makeTestVideo(t *testing.T, seconds float64, withAudio bool, audioOffset float64) []byte {
	t.Helper()
	requireFFmpeg(t)
	dir := t.TempDir()
	dur := strconv.FormatFloat(seconds, 'f', 3, 64)
	run := func(args ...string) {
		t.Helper()
		if b, err := exec.Command("ffmpeg", append([]string{"-v", "error", "-y"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("ffmpeg %v: %v: %s", args, err, b)
		}
	}

	videoOnly := filepath.Join(dir, "v.mp4")
	run("-f", "lavfi", "-i", "testsrc=size=64x64:rate=10:duration="+dur, "-pix_fmt", "yuv420p", videoOnly)
	out := videoOnly

	if withAudio {
		audioOnly := filepath.Join(dir, "a.m4a")
		run("-f", "lavfi", "-i", "sine=frequency=440:duration="+dur, "-c:a", "aac", audioOnly)

		// The offset is applied while remuxing with -c copy. Applied during
		// encoding, ffmpeg 4.4 (CI) and 6.1 (Docker image) fill the gap by
		// duplicating frames, leaving a file with no offset at all.
		offset := strconv.FormatFloat(max(audioOffset, -audioOffset), 'f', 3, 64)
		videoIn := []string{"-i", videoOnly}
		audioIn := []string{"-i", audioOnly}
		switch {
		case audioOffset > 0:
			audioIn = append([]string{"-itsoffset", offset}, audioIn...)
		case audioOffset < 0:
			videoIn = append([]string{"-itsoffset", offset}, videoIn...)
		}

		out = filepath.Join(dir, "in.mp4")
		args := append(videoIn, audioIn...)
		run(append(args, "-map", "0:v", "-map", "1:a", "-c", "copy", out)...)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func leadingSilence(t *testing.T, audio []byte) time.Duration {
	t.Helper()
	f := filepath.Join(t.TempDir(), "a.ogg")
	if err := os.WriteFile(f, audio, 0o644); err != nil {
		t.Fatal(err)
	}
	out, _ := exec.Command("ffmpeg", "-i", f, "-af", "silencedetect=n=-50dB:d=0.05", "-f", "null", "-").CombinedOutput()
	m := regexp.MustCompile(`silence_start: (-?[\d.]+)[\s\S]*?silence_end: ([\d.]+)`).FindSubmatch(out)
	if m == nil {
		return 0
	}
	start, _ := strconv.ParseFloat(string(m[1]), 64)
	if start > 0.05 {
		return 0
	}
	end, _ := strconv.ParseFloat(string(m[2]), 64)
	return time.Duration(end * float64(time.Second))
}

func assertNear(t *testing.T, name string, got, want, tolerance time.Duration) {
	t.Helper()
	if diff := got - want; diff < -tolerance || diff > tolerance {
		t.Errorf("%s = %v, want %v ± %v", name, got, want, tolerance)
	}
}

func TestExtractAudioCapsLengthAndEncodesOpus(t *testing.T) {
	video := makeTestVideo(t, AudioStickerMaxSeconds+3, true, 0)

	audio, length, err := ExtractAudio(video, 0)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "length", length, AudioStickerMaxSeconds*time.Second, 100*time.Millisecond)
	if len(audio) < 4 || string(audio[:4]) != "OggS" {
		t.Errorf("expected Ogg container, got header %q", audio[:min(4, len(audio))])
	}
}

func TestExtractAudioShortVideo(t *testing.T) {
	video := makeTestVideo(t, 2.5, true, 0)

	_, length, err := ExtractAudio(video, 0)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "length", length, 2500*time.Millisecond, 100*time.Millisecond)
}

func TestExtractAudioPadsLateAudio(t *testing.T) {
	video := makeTestVideo(t, 3, true, 0.4)

	audio, _, err := ExtractAudio(video, 0)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "leading silence", leadingSilence(t, audio), 400*time.Millisecond, 80*time.Millisecond)
}

func TestExtractAudioTrimsEarlyAudio(t *testing.T) {
	video := makeTestVideo(t, 3, true, -0.4)

	audio, length, err := ExtractAudio(video, 0)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "length", length, 2600*time.Millisecond, 100*time.Millisecond)
	if s := leadingSilence(t, audio); s > 50*time.Millisecond {
		t.Errorf("unexpected leading silence %v", s)
	}
}

func TestExtractAudioNoAudioTrack(t *testing.T) {
	video := makeTestVideo(t, 2, false, 0)

	if _, _, err := ExtractAudio(video, 0); !errors.Is(err, ErrNoAudio) {
		t.Fatalf("got %v, want ErrNoAudio", err)
	}
}

// animationLength sums the frame durations of an animated webp.
func animationLength(t *testing.T, webp []byte) time.Duration {
	t.Helper()
	var total int
	for i := 12; i+8 <= len(webp); {
		size := int(binary.LittleEndian.Uint32(webp[i+4 : i+8]))
		if string(webp[i:i+4]) == "ANMF" {
			d := webp[i+8+12 : i+8+15]
			total += int(d[0]) | int(d[1])<<8 | int(d[2])<<16
		}
		i += 8 + size + size&1
	}
	return time.Duration(total) * time.Millisecond
}

func TestAudioStickerAnimationMatchesAudio(t *testing.T) {
	if _, err := exec.LookPath("webpmux"); err != nil {
		t.Skip("webpmux not available")
	}
	video := makeTestVideo(t, 7.3, true, 0)

	_, length, err := ExtractAudio(video, 0)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	st := New("a", "b", video, StickerSquash)
	st.Length = length
	webp, err := st.Convert()
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	// One frame of tolerance at the lowest fps of the ladder.
	assertNear(t, "animation length", animationLength(t, webp), length, 125*time.Millisecond)
}

func TestExtractAudioCustomMaxLength(t *testing.T) {
	video := makeTestVideo(t, 8, true, 0)

	_, length, err := ExtractAudio(video, 3*time.Second)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "length", length, 3*time.Second, 100*time.Millisecond)
}

// oggGranules returns the granule position of each Ogg page.
func oggGranules(t *testing.T, data []byte) []int64 {
	t.Helper()
	var granules []int64
	for i := 0; i+27 <= len(data); {
		if string(data[i:i+4]) != "OggS" {
			t.Fatalf("bad Ogg page at %d", i)
		}
		granules = append(granules, int64(binary.LittleEndian.Uint64(data[i+6:i+14])))
		segments := int(data[i+26])
		size := 0
		for _, s := range data[i+27 : i+27+segments] {
			size += int(s)
		}
		i += 27 + segments + size
	}
	return granules
}

// TestExtractAudioGranulesOnFrameGrid guards the fix for silent audio
// stickers: WhatsApp only plays sticker audio whose Ogg granule positions sit
// on the 960-sample (20ms at 48kHz) Opus frame grid. A 44.1kHz source with an
// offset is the case that used to produce misaligned granules.
func TestExtractAudioGranulesOnFrameGrid(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "src.mp4")
	if b, err := exec.Command("ffmpeg", "-v", "error", "-y",
		"-f", "lavfi", "-i", "testsrc=size=64x64:rate=30:duration=5",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=44100:duration=5",
		"-c:a", "aac", "-pix_fmt", "yuv420p", src).CombinedOutput(); err != nil {
		t.Fatalf("generate video: %v: %s", err, b)
	}
	video, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}

	for _, video := range [][]byte{video, makeTestVideo(t, 3, true, 0.4), makeTestVideo(t, 3, true, -0.4)} {
		audio, _, err := ExtractAudio(video, 0)
		if err != nil {
			t.Fatalf("ExtractAudio: %v", err)
		}
		granules := oggGranules(t, audio)
		// Skip the two header pages and the last page (end trimming).
		for i, g := range granules[2 : len(granules)-1] {
			if g%960 != 0 {
				t.Errorf("page %d granule %d is not a multiple of 960 (all: %v)", i+2, g, granules)
			}
		}
	}
}
