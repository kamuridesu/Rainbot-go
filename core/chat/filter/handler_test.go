package filter

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
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

func encodeTestWebP(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	webpBytes, err := utils.ToWebp(buf.Bytes())
	if err != nil {
		t.Fatalf("failed to convert test image to webp: %v", err)
	}
	return webpBytes
}

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
		Filter: services.NewFilterRepository(repositories.NewFilterRepository(sqlDB)),
	}
}

func newTestMessage(t *testing.T, fake *botfakes.FakeClient, db *database.DatabaseSingleton, text string) *messages.Message {
	t.Helper()

	chatJID := types.NewJID("123456", types.GroupServer)
	if _, err := db.Chat.GetOrCreateChat(chatJID.String()); err != nil {
		t.Fatalf("failed to seed chat: %v", err)
	}
	chat, err := db.Chat.Get(chatJID.String())
	if err != nil || chat == nil {
		t.Fatalf("failed to fetch seeded chat: %v", err)
	}

	return &messages.Message{
		Ctx:  context.Background(),
		Bot:  &bot.Bot{Client: fake, DB: db},
		Text: &text,
		Chat: chat,
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{Chat: chatJID, Sender: types.NewJID("111", types.HiddenUserServer)},
				ID:            "stanza-1",
			},
		},
	}
}

func TestGetChatFiltersNoMatch(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, "hello world")

	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: "text", Response: "ola"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	if len(fake.SentMessages) != 0 {
		t.Errorf("expected no reply for a non-matching message, got %v", fake.SentMessages)
	}
}

func TestGetChatFiltersTextMatch(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, "diga oi por favor")

	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: "text", Response: "ola!"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[0].Message.GetExtendedTextMessage().GetText() != "ola!" {
		t.Errorf("reply text = %q, want %q", fake.SentMessages[0].Message.GetExtendedTextMessage().GetText(), "ola!")
	}
}

func TestGetChatFiltersMediaMatchMissingFile(t *testing.T) {
	db := newTestDB(t)
	t.Chdir(t.TempDir())
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, "manda a figura")

	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "figura", Kind: "sticker", Response: "does-not-exist.webp"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + missing-file reply), got %d", len(fake.SentMessages))
	}
}

func TestGetChatFiltersMediaMatchSuccess(t *testing.T) {
	db := newTestDB(t)
	t.Chdir(t.TempDir())
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, "manda a figura")

	webpBytes := encodeTestWebP(t)
	fileID := storage.RandomFilename("webp")
	if err := storage.NewFile(fileID).Write(m.Ctx, webpBytes); err != nil {
		t.Fatalf("seed file error = %v", err)
	}
	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "figura", Kind: "sticker", Response: fileID}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	if len(fake.SentMessages) != 1 || fake.SentMessages[0].Message.GetStickerMessage() == nil {
		t.Errorf("expected a sticker reply, got %v", fake.SentMessages)
	}
}
