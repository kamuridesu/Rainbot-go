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
	dur := strconv.FormatFloat(seconds, 'f', 3, 64)
	out := filepath.Join(t.TempDir(), "in.mp4")

	var args []string
	if audioOffset < 0 {
		args = append(args, "-itsoffset", strconv.FormatFloat(-audioOffset, 'f', 3, 64))
	}
	args = append(args, "-f", "lavfi", "-i", "testsrc=size=64x64:rate=10:duration="+dur)
	if withAudio {
		if audioOffset > 0 {
			args = append(args, "-itsoffset", strconv.FormatFloat(audioOffset, 'f', 3, 64))
		}
		args = append(args, "-f", "lavfi", "-i", "sine=frequency=440:duration="+dur)
	}
	args = append(args, "-v", "error", "-pix_fmt", "yuv420p", out)
	if b, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("generate video: %v: %s", err, b)
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

	audio, length, err := ExtractAudio(video)
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

	_, length, err := ExtractAudio(video)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "length", length, 2500*time.Millisecond, 100*time.Millisecond)
}

func TestExtractAudioPadsLateAudio(t *testing.T) {
	video := makeTestVideo(t, 3, true, 0.4)

	audio, _, err := ExtractAudio(video)
	if err != nil {
		t.Fatalf("ExtractAudio: %v", err)
	}
	assertNear(t, "leading silence", leadingSilence(t, audio), 400*time.Millisecond, 80*time.Millisecond)
}

func TestExtractAudioTrimsEarlyAudio(t *testing.T) {
	video := makeTestVideo(t, 3, true, -0.4)

	audio, length, err := ExtractAudio(video)
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

	if _, _, err := ExtractAudio(video); !errors.Is(err, ErrNoAudio) {
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

	_, length, err := ExtractAudio(video)
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
