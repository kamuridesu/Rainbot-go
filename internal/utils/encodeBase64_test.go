package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncode64(t *testing.T) {
	data := []byte("hello world")
	want := base64.StdEncoding.EncodeToString(data)

	t.Run("without data prefix", func(t *testing.T) {
		got := Encode64(data, false)
		if got != want {
			t.Errorf("Encode64() = %q, want %q", got, want)
		}
	})

	t.Run("with data prefix", func(t *testing.T) {
		got := Encode64(data, true)
		wantPrefixed := "data:image/jpg;base64," + want
		if got != wantPrefixed {
			t.Errorf("Encode64() = %q, want %q", got, wantPrefixed)
		}
		if !strings.HasPrefix(got, "data:image/jpg;base64,") {
			t.Errorf("Encode64() missing expected prefix, got %q", got)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		if got := Encode64(nil, false); got != "" {
			t.Errorf("Encode64(nil) = %q, want empty string", got)
		}
	})
}
