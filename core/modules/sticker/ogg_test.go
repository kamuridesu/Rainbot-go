package sticker

import (
	"errors"
	"testing"
	"time"
)

func TestOggOpusDurationMatchesExtractAudio(t *testing.T) {
	for _, offset := range []float64{0, 0.4, -0.4} {
		video := makeTestVideo(t, 3, true, offset)
		audio, length, err := ExtractAudio(video, 0)
		if err != nil {
			t.Fatalf("ExtractAudio: %v", err)
		}
		got, err := OggOpusDuration(audio)
		if err != nil {
			t.Fatalf("OggOpusDuration: %v", err)
		}
		assertNear(t, "duration", got, length, 50*time.Millisecond)
	}
}

func TestOggOpusDurationRejectsInvalid(t *testing.T) {
	for name, data := range map[string][]byte{
		"empty":     nil,
		"not ogg":   []byte("RIFF....WEBPVP8 "),
		"truncated": []byte("OggS\x00\x02"),
	} {
		if _, err := OggOpusDuration(data); !errors.Is(err, ErrInvalidOgg) {
			t.Errorf("%s: got %v, want ErrInvalidOgg", name, err)
		}
	}
}

func TestDurationSeconds(t *testing.T) {
	for d, want := range map[time.Duration]uint32{
		0:                       1,
		300 * time.Millisecond:  1,
		2400 * time.Millisecond: 2,
		2600 * time.Millisecond: 3,
		10 * time.Second:        10,
	} {
		if got := DurationSeconds(d); got != want {
			t.Errorf("DurationSeconds(%v) = %d, want %d", d, got, want)
		}
	}
}
