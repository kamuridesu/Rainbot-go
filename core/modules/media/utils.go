package media

import "fmt"

type MediaType = string

var (
	MediaVideo MediaType = "video"
	MediaAudio MediaType = "audio"
)

func readLittleEndianInt(data []byte) int64 {
	return int64(uint32(data[0]) | uint32(data[1])<<8 | uint32(data[2])<<16 | uint32(data[3])<<24)
}

func GetOggDurationMs(data []byte) (int64, error) {

	var length int64
	for i := len(data) - 14; i >= 0 && length == 0; i-- {
		if data[i] == 'O' && data[i+1] == 'g' && data[i+2] == 'g' && data[i+3] == 'S' {
			length = int64(readLittleEndianInt(data[i+6 : i+14]))
		}
	}

	var rate int64
	for i := 0; i < len(data)-14 && rate == 0; i++ {
		if data[i] == 'v' && data[i+1] == 'o' && data[i+2] == 'r' && data[i+3] == 'b' && data[i+4] == 'i' && data[i+5] == 's' {
			rate = int64(readLittleEndianInt(data[i+11 : i+15]))
		}
	}

	if length == 0 || rate == 0 {
		return 0, fmt.Errorf("could not find necessary information in Ogg file")
	}

	durationMs := length * 1000 / rate
	return durationMs, nil

}
