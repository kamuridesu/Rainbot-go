package fun

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func requireStickerTools(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available, skipping sticker conversion test")
	}
	if _, err := exec.LookPath("webpmux"); err != nil {
		t.Skip("webpmux not available, skipping sticker conversion test")
	}
}

func testPNGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	return buf.Bytes()
}

func TestNewStickerNoMediaFound(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Type = messages.TextMessage

	NewStickerSquash(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + fail reaction + rejection), got %d", len(fake.SentMessages))
	}
}

func TestNewStickerFromCurrentMessage(t *testing.T) {
	requireStickerTools(t)

	png := testPNGBytes(t)
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return png, nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Type = messages.ImageMessage
	m.RawMessage = &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}}

	NewStickerSquash(m)

	// Waiting reaction, then ReplySticker's own (success reaction + sticker).
	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + success reaction + sticker), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[2].Message.GetStickerMessage() == nil {
		t.Errorf("expected a sticker message to be sent, got %+v", fake.SentMessages[2])
	}
}

func TestNewStickerFromQuotedMessage(t *testing.T) {
	requireStickerTools(t)

	png := testPNGBytes(t)
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return png, nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Type = messages.TextMessage
	m.QuotedMessage = &messages.Message{
		Ctx:        m.Ctx,
		Bot:        m.Bot,
		Type:       messages.ImageMessage,
		RawMessage: &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}},
	}

	NewStickerOriginal(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + success reaction + sticker), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[2].Message.GetStickerMessage() == nil {
		t.Error("expected a sticker message to be sent from the quoted message's media")
	}
}

func TestNewStickerDownloadError(t *testing.T) {
	wantErr := errors.New("download failed")
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return nil, wantErr
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Type = messages.ImageMessage
	m.RawMessage = &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}}

	NewStickerSquash(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + fail reaction + error reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[2].Message.GetExtendedTextMessage().GetText()
	if text == "" {
		t.Error("expected an error reply describing the download failure")
	}
}

func TestNewAudioStickerRequiresVideo(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Type = messages.ImageMessage

	NewAudioSticker(m)

	// Fail reaction + rejection reply, no sticker.
	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages, got %d", len(fake.SentMessages))
	}
}

func TestNewAudioStickerFromQuotedVideo(t *testing.T) {
	requireStickerTools(t)
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available")
	}

	out := t.TempDir() + "/in.mp4"
	if b, err := exec.Command("ffmpeg", "-v", "error",
		"-f", "lavfi", "-i", "testsrc=size=64x64:rate=10:duration=2",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-pix_fmt", "yuv420p", out).CombinedOutput(); err != nil {
		t.Fatalf("generate video: %v: %s", err, b)
	}
	video, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}

	var uploaded []whatsmeow.MediaType
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return video, nil
		},
		UploadFunc: func(ctx context.Context, data []byte, mt whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
			uploaded = append(uploaded, mt)
			return whatsmeow.UploadResponse{URL: "u", DirectPath: "/d", FileLength: uint64(len(data))}, nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Type = messages.TextMessage
	m.QuotedMessage = &messages.Message{
		Ctx:        m.Ctx,
		Bot:        m.Bot,
		Type:       messages.VideoMessage,
		RawMessage: &waE2E.Message{VideoMessage: &waE2E.VideoMessage{}},
	}

	NewAudioSticker(m)

	if len(uploaded) != 2 || uploaded[0] != whatsmeow.MediaImage || uploaded[1] != whatsmeow.MediaAudio {
		t.Fatalf("unexpected uploads: %v", uploaded)
	}
	last := fake.SentMessages[len(fake.SentMessages)-1].Message
	stk := last.GetAudioStickerMessage().GetMessage().GetStickerMessage()
	if stk == nil {
		t.Fatalf("expected AudioStickerMessage wrapper, got %+v", last)
	}
	if !stk.GetIsAnimated() || stk.GetMimetype() != "image/webp" {
		t.Errorf("unexpected sticker fields: animated=%v mimetype=%q", stk.GetIsAnimated(), stk.GetMimetype())
	}
	outer := last.GetMessageContextInfo().GetMessageSecret()
	inner := last.GetAudioStickerMessage().GetMessage().GetMessageContextInfo().GetMessageSecret()
	if len(outer) != 32 || !bytes.Equal(outer, inner) {
		t.Errorf("expected same 32-byte messageSecret outside and inside the wrapper, got %x / %x", outer, inner)
	}
	if stk.GetWidth() != 512 || stk.GetHeight() != 512 {
		t.Errorf("size = %dx%d, want 512x512", stk.GetWidth(), stk.GetHeight())
	}
	aud := stk.GetAudioMessage()
	if aud.GetMimetype() != "audio/ogg; codecs=opus" || aud.GetSeconds() != 2 {
		t.Errorf("unexpected audio fields: mimetype=%q seconds=%d", aud.GetMimetype(), aud.GetSeconds())
	}
}
