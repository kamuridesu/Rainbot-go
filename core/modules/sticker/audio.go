package sticker

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const AudioStickerMaxSeconds = 10

var ErrNoAudio = errors.New("video has no audio track")

func ExtractAudio(video []byte) ([]byte, time.Duration, error) {
	input, err := CreateTempFile(video)
	if err != nil {
		return nil, 0, err
	}
	defer DeleteTmpFile(input)

	hasAudio, err := hasAudioStream(input)
	if err != nil {
		return nil, 0, err
	}
	if !hasAudio {
		return nil, 0, ErrNoAudio
	}

	output := input + "-audio.ogg"
	defer DeleteTmpFile(output)

	args := []string{"-i", input}
	if filter := alignFilter(input); filter != "" {
		args = append(args, "-af", filter)
	}
	args = append(args,
		"-t", strconv.Itoa(AudioStickerMaxSeconds),
		"-vn",
		"-map", "0:a:0",
		"-c:a", "libopus",
		"-b:a", "64k",
		"-ac", "1",
		"-f", "ogg",
		output,
	)
	if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		slog.Error("FFmpeg audio extraction failed", "error", err.Error(), "output", string(out))
		return nil, 0, fmt.Errorf("ffmpeg error: %w", err)
	}

	duration, err := probeDuration(output)
	if err != nil {
		return nil, 0, err
	}

	data, err := os.ReadFile(output)
	if err != nil {
		return nil, 0, err
	}
	return data, duration, nil
}

func alignFilter(input string) string {
	videoStart, errV := probeStartTime(input, "v:0")
	audioStart, errA := probeStartTime(input, "a:0")
	if errV != nil || errA != nil {
		return ""
	}

	offset := audioStart - videoStart
	switch {
	case offset > 0.01:
		return fmt.Sprintf("adelay=%d:all=1", int(math.Round(offset*1000)))
	case offset < -0.01:
		return fmt.Sprintf("atrim=start=%.3f,asetpts=PTS-STARTPTS", -offset)
	default:
		return ""
	}
}

func probeStartTime(input, stream string) (float64, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", stream,
		"-show_entries", "stream=start_time",
		"-of", "csv=p=0",
		input,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe error: %w", err)
	}
	return strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
}

func hasAudioStream(input string) (bool, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		input,
	).Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe error: %w", err)
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func probeDuration(input string) (time.Duration, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		input,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe error: %w", err)
	}
	secs, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", out, err)
	}
	return time.Duration(secs * float64(time.Second)), nil
}
