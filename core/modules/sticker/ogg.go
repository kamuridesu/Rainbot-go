package sticker

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

var ErrInvalidOgg = errors.New("invalid ogg opus stream")

func OggOpusDuration(data []byte) (time.Duration, error) {
	var preSkip uint16
	var lastGranule int64 = -1
	for i, page := 0, 0; i < len(data); page++ {
		if i+27 > len(data) || string(data[i:i+4]) != "OggS" {
			return 0, ErrInvalidOgg
		}
		segments := int(data[i+26])
		bodyStart := i + 27 + segments
		if bodyStart > len(data) {
			return 0, ErrInvalidOgg
		}
		size := 0
		for _, s := range data[i+27 : bodyStart] {
			size += int(s)
		}
		if bodyStart+size > len(data) {
			return 0, ErrInvalidOgg
		}
		body := data[bodyStart : bodyStart+size]
		if page == 0 {
			if len(body) < 19 || string(body[:8]) != "OpusHead" {
				return 0, ErrInvalidOgg
			}
			preSkip = binary.LittleEndian.Uint16(body[10:12])
		}
		if g := int64(binary.LittleEndian.Uint64(data[i+6 : i+14])); g >= 0 {
			lastGranule = g
		}
		i = bodyStart + size
	}
	samples := lastGranule - int64(preSkip)
	if samples <= 0 {
		return 0, ErrInvalidOgg
	}
	return time.Duration(samples) * time.Second / 48000, nil
}

func DurationSeconds(d time.Duration) uint32 {
	return uint32(max(1, math.Round(d.Seconds())))
}
