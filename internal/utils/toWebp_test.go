package utils

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func encodeTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := range 4 {
		for x := range 4 {
			img.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	return buf.Bytes()
}

func TestToWebpConvertsValidImage(t *testing.T) {
	png := encodeTestPNG(t)

	out, err := ToWebp(png)
	if err != nil {
		t.Fatalf("ToWebp() error = %v", err)
	}

	if len(out) < 12 {
		t.Fatalf("ToWebp() output too short to be a valid webp: %d bytes", len(out))
	}
	if string(out[0:4]) != "RIFF" || string(out[8:12]) != "WEBP" {
		t.Errorf("ToWebp() output missing RIFF/WEBP header, got header %q/%q", out[0:4], out[8:12])
	}
}

func TestToWebpRejectsInvalidImage(t *testing.T) {
	_, err := ToWebp([]byte("not an image"))
	if err == nil {
		t.Fatal("expected an error for non-image input")
	}
}
