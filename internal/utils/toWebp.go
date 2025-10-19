package utils

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/chai2010/webp"
	"google.golang.org/protobuf/proto"
)

func ToWebp(media []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(media))
	if err != nil {
		return nil, err
	}
	return webp.EncodeRGBA(img, *proto.Float32(1))
}
