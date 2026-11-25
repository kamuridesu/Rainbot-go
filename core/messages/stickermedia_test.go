package messages

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/util/cbcutil"
	"google.golang.org/protobuf/proto"
)

func chunk(tag string, payload []byte) []byte {
	b := append([]byte(tag), binary.LittleEndian.AppendUint32(nil, uint32(len(payload)))...)
	b = append(b, payload...)
	if len(payload)%2 == 1 {
		b = append(b, 0)
	}
	return b
}

func buildWebp(chunks ...[]byte) []byte {
	body := []byte("WEBP")
	for _, c := range chunks {
		body = append(body, c...)
	}
	return append(append([]byte("RIFF"), binary.LittleEndian.AppendUint32(nil, uint32(len(body)))...), body...)
}

func TestWebpImageDataLength(t *testing.T) {
	frames := [][]byte{chunk("VP8X", make([]byte, 10)), chunk("ANIM", make([]byte, 6)), chunk("ANMF", make([]byte, 33))}
	noExif := buildWebp(frames...)
	withExif := buildWebp(append(frames, chunk("EXIF", []byte("metadata")))...)

	if got, err := webpImageDataLength(withExif); err != nil || got != len(noExif) {
		t.Errorf("with EXIF: got %d (%v), want %d", got, err, len(noExif))
	}
	if got, err := webpImageDataLength(noExif); err != nil || got != len(noExif) {
		t.Errorf("without EXIF: got %d (%v), want %d", got, err, len(noExif))
	}
	if _, err := webpImageDataLength([]byte("not a webp")); err == nil {
		t.Error("expected error for non-webp input")
	}
}

// TestFirstFrameSidecarMatchesUploadEncryption checks the sidecar against the
// ciphertext whatsmeow itself produces on upload, so a mismatch between our
// re-encryption and the real upload would be caught.
func TestFirstFrameSidecarMatchesUploadEncryption(t *testing.T) {
	mediaKey := make([]byte, 32)
	rand.Read(mediaKey)
	iv, cipherKey, macKey, err := mediaKeys(mediaKey, whatsmeow.MediaImage)
	if err != nil {
		t.Fatal(err)
	}

	for _, size := range []int{100, 4096, 4099} {
		plaintext := make([]byte, size)
		rand.Read(plaintext)
		ciphertext, err := cbcutil.Encrypt(cipherKey, iv, plaintext)
		if err != nil {
			t.Fatal(err)
		}
		for _, length := range []int{1, 16, 17, size - 40, size} {
			end := min((length+15)/16*16, len(ciphertext))
			mac := hmac.New(sha256.New, macKey)
			mac.Write(iv)
			mac.Write(ciphertext[:end])
			want := mac.Sum(nil)[:10]

			got, err := firstFrameSidecar(plaintext, mediaKey, length)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("size=%d length=%d: got %x, want %x", size, length, got, want)
			}
		}
	}
}

func TestWithE2EEMediaDomain(t *testing.T) {
	key := []byte{1, 2, 3}

	fresh := withE2EEMediaDomain(nil, key)
	if fresh.GetMediaDomainInfo().GetMediaKeyDomain() != waE2E.MediaKeyDomain_MEDIA_KEY_DOMAIN_E2EE || !bytes.Equal(fresh.GetMediaDomainInfo().GetE2EeMediaKey(), key) {
		t.Errorf("unexpected media domain info: %+v", fresh.GetMediaDomainInfo())
	}

	quote := &waE2E.ContextInfo{StanzaID: proto.String("abc")}
	merged := withE2EEMediaDomain(quote, key)
	if merged.GetStanzaID() != "abc" || merged.GetMediaDomainInfo() == nil {
		t.Errorf("quote info lost or domain missing: %+v", merged)
	}
	if quote.MediaDomainInfo != nil {
		t.Error("input context was mutated")
	}
}
