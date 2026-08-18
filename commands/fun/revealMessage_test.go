package fun

import (
	"context"
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func TestRevealMessageImage(t *testing.T) {
	var gotMsg whatsmeow.DownloadableMessage
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			gotMsg = msg
			return []byte("image-bytes"), nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	img := &waE2E.ImageMessage{}
	m.QuotedMessage = &messages.Message{
		Type:       messages.ImageMessage,
		RawMessage: &waE2E.Message{ImageMessage: img},
	}

	RevealMessage(m)

	if gotMsg != img {
		t.Errorf("expected Download() to be called with the quoted image message")
	}
	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (media + success reaction), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[0].Message.GetImageMessage() == nil {
		t.Error("expected an image message to be sent")
	}
}

func TestRevealMessageVideo(t *testing.T) {
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return []byte("video-bytes"), nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.QuotedMessage = &messages.Message{
		Type:       messages.VideoMessage,
		RawMessage: &waE2E.Message{VideoMessage: &waE2E.VideoMessage{}},
	}

	RevealMessage(m)

	if len(fake.SentMessages) != 2 || fake.SentMessages[0].Message.GetVideoMessage() == nil {
		t.Errorf("expected a video message to be sent, got %v", fake.SentMessages)
	}
}

func TestRevealMessageDownloadError(t *testing.T) {
	wantErr := errors.New("download failed")
	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return nil, wantErr
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.QuotedMessage = &messages.Message{
		Type:       messages.ImageMessage,
		RawMessage: &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}},
	}

	RevealMessage(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + error reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if text == "" {
		t.Error("expected an error reply")
	}
}
