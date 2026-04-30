package utils

import (
	"encoding/base64"
	"fmt"
)

func Encode64(data []byte, withDataPrefix bool) string {
	enc := base64.StdEncoding.EncodeToString(data)
	if withDataPrefix {
		return fmt.Sprintf("data:image/jpg;base64,%s", enc)
	}
	return enc
}
