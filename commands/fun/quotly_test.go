package fun

import (
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"github.com/kamuridesu/rainbot-go/internal/storage"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func setQuotedStanzaID(m *messages.Message, stanzaID string) {
	m.RawMessage = &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			ContextInfo: &waE2E.ContextInfo{StanzaID: proto.String(stanzaID)},
		},
	}
}

func TestRandomQuoteNoQuotesSaved(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	RandomQuote(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (fail reaction), got %d", len(fake.SentMessages))
	}
}

func TestRandomQuoteSuccess(t *testing.T) {
	db := newTestDB(t)

	// storage reads/writes a bare filename, resolved relative to the CWD.
	// Must run after newTestDB(t), which itself chdirs to the repo root to
	// find the migration files.
	t.Chdir(t.TempDir())

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)

	fileID := storage.RandomFilename("png")
	if err := storage.NewFile(fileID).Write(m.Ctx, []byte("sticker-bytes")); err != nil {
		t.Fatalf("seed file write error = %v", err)
	}
	if err := db.Quotly.SaveQuotly(&models.QuotlyFile{ChatID: m.Chat.ChatID, FileId: fileID}); err != nil {
		t.Fatalf("seed SaveQuotly error = %v", err)
	}

	RandomQuote(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + sticker), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[1].Message.GetStickerMessage() == nil {
		t.Error("expected a sticker message to be sent")
	}

	sent, err := db.Quotly.GetSentMessageByStanzaID(m.Chat.ChatID, "")
	if err != nil {
		t.Fatalf("GetSentMessageByStanzaID() error = %v", err)
	}
	if sent == nil || sent.FileId != fileID {
		t.Errorf("expected the sent-message mapping to be recorded for %q, got %+v", fileID, sent)
	}
}

func TestHandleQuoteDeleteCommandNoStanza(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	HandleQuoteDeleteCommand(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + rejection), got %d", len(fake.SentMessages))
	}
}

func TestHandleQuoteDeleteCommandUnknownStanza(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	setQuotedStanzaID(m, "unknown-stanza")

	HandleQuoteDeleteCommand(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + rejection), got %d", len(fake.SentMessages))
	}
}

func TestHandleQuoteDeleteCommandSuccess(t *testing.T) {
	db := newTestDB(t)
	t.Chdir(t.TempDir())
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)

	fileID := storage.RandomFilename("png")
	if err := storage.NewFile(fileID).Write(m.Ctx, []byte("sticker-bytes")); err != nil {
		t.Fatalf("seed file write error = %v", err)
	}
	if err := db.Quotly.SaveQuotly(&models.QuotlyFile{ChatID: m.Chat.ChatID, FileId: fileID}); err != nil {
		t.Fatalf("seed SaveQuotly error = %v", err)
	}
	if err := db.Quotly.SaveSentMessage(&models.QuotlyMessage{
		StanzaID:  "the-stanza-id",
		ChatID:    m.Chat.ChatID,
		FileId:    fileID,
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("seed SaveSentMessage error = %v", err)
	}
	setQuotedStanzaID(m, "the-stanza-id")

	HandleQuoteDeleteCommand(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (success reaction), got %d", len(fake.SentMessages))
	}

	remaining, err := db.Quotly.GetAllByChat(m.Chat.ChatID)
	if err != nil {
		t.Fatalf("GetAllByChat() error = %v", err)
	}
	if len(remaining) != 0 {
		t.Errorf("expected the quote to be deleted, got %+v", remaining)
	}

	if exists, _ := storage.NewFile(fileID).Exists(m.Ctx); exists {
		t.Error("expected the sticker file to be deleted from storage")
	}
}
