package quotly

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database"
	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/providers"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/services"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"github.com/kamuridesu/rainbot-go/internal/storage"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func newTestDB(t *testing.T) *database.DatabaseSingleton {
	t.Helper()

	dbfile := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := providers.InitDB("sqlite3", "file:"+dbfile+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	_, file, _, _ := runtime.Caller(0)
	t.Chdir(filepath.Join(filepath.Dir(file), "..", "..", ".."))

	if err := utils.MigrateSqlite(sqlDB); err != nil {
		t.Fatalf("MigrateSqlite() error = %v", err)
	}

	return &database.DatabaseSingleton{
		Chat:   services.NewChatService(repositories.NewChatRepository(sqlDB)),
		Quotly: services.NewQuotlyService(repositories.NewQuotlyRepository(sqlDB)),
	}
}

func newTestMessage(t *testing.T, fake *botfakes.FakeClient, db *database.DatabaseSingleton, chat *models.Chat) *messages.Message {
	t.Helper()
	chatJID := types.NewJID("123456", types.GroupServer)
	chat.ChatID = chatJID.String()
	if _, err := db.Chat.GetOrCreateChat(chatJID.String()); err != nil {
		t.Fatalf("seed chat error = %v", err)
	}

	return &messages.Message{
		Ctx:  context.Background(),
		Bot:  &bot.Bot{Client: fake, DB: db},
		Chat: chat,
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{Chat: chatJID},
				ID:            "stanza-1",
			},
		},
	}
}

func TestRandomQuoteDropDisabledForChat(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, &models.Chat{AllowQuote: 0})

	RandomQuoteDrop(m)

	if len(fake.SentMessages) != 0 {
		t.Error("expected no message when quoting is disabled for the chat")
	}
}

func TestRandomQuoteDropNoQuotesSaved(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, &models.Chat{AllowQuote: 1, QuoteNMessages: 1})

	RandomQuoteDrop(m)

	if len(fake.SentMessages) != 0 {
		t.Error("expected no message when there are no saved quotes for the chat")
	}
}

func TestRandomQuoteDropMissingFile(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, &models.Chat{AllowQuote: 1, QuoteNMessages: 1})

	if err := db.Quotly.SaveQuotly(&models.QuotlyFile{ChatID: m.Chat.ChatID, FileId: "does-not-exist.png"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	RandomQuoteDrop(m)

	if len(fake.SentMessages) != 0 {
		t.Error("expected no message when the stored quote file is missing from storage")
	}
}

func TestRandomQuoteDropSuccess(t *testing.T) {
	db := newTestDB(t)
	t.Chdir(t.TempDir())
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, &models.Chat{AllowQuote: 1, QuoteNMessages: 1})

	fileID := storage.RandomFilename("png")
	if err := storage.NewFile(fileID).Write(m.Ctx, []byte("sticker-bytes")); err != nil {
		t.Fatalf("seed file error = %v", err)
	}
	if err := db.Quotly.SaveQuotly(&models.QuotlyFile{ChatID: m.Chat.ChatID, FileId: fileID}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	RandomQuoteDrop(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (the sticker), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[0].Message.GetStickerMessage() == nil {
		t.Error("expected a sticker message to be sent")
	}
}
