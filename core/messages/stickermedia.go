package messages

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func webpImageDataLength(webp []byte) (int, error) {
	if len(webp) < 12 || string(webp[:4]) != "RIFF" || string(webp[8:12]) != "WEBP" {
		return 0, errors.New("not a webp file")
	}
	for i := 12; i+8 <= len(webp); {
		if string(webp[i:i+4]) == "EXIF" {
			return i, nil
		}
		size := int(binary.LittleEndian.Uint32(webp[i+4 : i+8]))
		i += 8 + size + size&1
	}
	return len(webp), nil
}

func mediaKeys(mediaKey []byte, appInfo whatsmeow.MediaType) (iv, cipherKey, macKey []byte, err error) {
	expanded, err := hkdf.Key(sha256.New, mediaKey, nil, string(appInfo), 112)
	if err != nil {
		return nil, nil, nil, err
	}
	return expanded[:16], expanded[16:48], expanded[48:80], nil
}

func firstFrameSidecar(plaintext, mediaKey []byte, length int) ([]byte, error) {
	iv, cipherKey, macKey, err := mediaKeys(mediaKey, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, err
	}

	pad := aes.BlockSize - len(plaintext)%aes.BlockSize
	padded := make([]byte, len(plaintext)+pad)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(pad)
	}
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)

	end := min((length+aes.BlockSize-1)/aes.BlockSize*aes.BlockSize, len(ciphertext))
	mac := hmac.New(sha256.New, macKey)
	mac.Write(iv)
	mac.Write(ciphertext[:end])
	return mac.Sum(nil)[:10], nil
}

func withE2EEMediaDomain(ctx *waE2E.ContextInfo, mediaKey []byte) *waE2E.ContextInfo {
	out := &waE2E.ContextInfo{}
	if ctx != nil {
		out = proto.Clone(ctx).(*waE2E.ContextInfo)
	}
	out.MediaDomainInfo = &waE2E.MediaDomainInfo{
		MediaKeyDomain: waE2E.MediaKeyDomain_MEDIA_KEY_DOMAIN_E2EE.Enum(),
		E2EeMediaKey:   mediaKey,
	}
	return out
}
