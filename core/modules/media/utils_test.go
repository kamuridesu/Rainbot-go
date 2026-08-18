package media

import (
	"encoding/binary"
	"testing"
)

func TestReadLittleEndianInt(t *testing.T) {
	data := []byte{0x01, 0x00, 0x00, 0x00}
	if got := readLittleEndianInt(data); got != 1 {
		t.Errorf("readLittleEndianInt() = %d, want 1", got)
	}

	data = []byte{0x00, 0xCA, 0x9A, 0x3B} // 1_000_000_000 little-endian
	if got := readLittleEndianInt(data); got != 1_000_000_000 {
		t.Errorf("readLittleEndianInt() = %d, want 1000000000", got)
	}
}

// buildTestOggVorbis constructs a minimal byte sequence that satisfies
// GetOggDurationMs's pattern search: a "vorbis" codec header carrying the
// sample rate, and an "OggS" page carrying the granule position (samples).
func buildTestOggVorbis(t *testing.T, granule, rate uint32) []byte {
	t.Helper()

	var buf []byte
	buf = append(buf, make([]byte, 5)...) // leading padding
	vorbisStart := len(buf)
	buf = append(buf, []byte("vorbis")...)
	// Sample rate must land at data[vorbisStart+11 : vorbisStart+15].
	for len(buf) < vorbisStart+11 {
		buf = append(buf, 0)
	}
	rateBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(rateBytes, rate)
	buf = append(buf, rateBytes...)
	buf = append(buf, make([]byte, 8)...) // trailing padding

	oggStart := len(buf)
	buf = append(buf, []byte("OggS")...)
	buf = append(buf, 0x00, 0x04) // version, header_type (unused by the parser)
	granuleBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(granuleBytes, granule)
	buf = append(buf, granuleBytes...)
	buf = append(buf, make([]byte, 8)...) // padding so data[oggStart+6:oggStart+14] is in bounds
	_ = oggStart

	return buf
}

func TestGetOggDurationMs(t *testing.T) {
	// 88200 samples at 44100 Hz = 2000ms.
	data := buildTestOggVorbis(t, 88200, 44100)

	got, err := GetOggDurationMs(data)
	if err != nil {
		t.Fatalf("GetOggDurationMs() error = %v", err)
	}
	if got != 2000 {
		t.Errorf("GetOggDurationMs() = %d, want 2000", got)
	}
}

func TestGetOggDurationMsMissingMarkers(t *testing.T) {
	if _, err := GetOggDurationMs([]byte("not an ogg file at all")); err == nil {
		t.Error("expected an error when neither OggS nor vorbis markers are present")
	}
}
