package filter

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
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
	"go.mau.fi/whatsmeow"
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

	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: KindText, Response: "ola"}); err != nil {
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

	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: KindText, Response: "ola!"}); err != nil {
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

	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "figura", Kind: KindSticker, Response: "does-not-exist.webp"}); err != nil {
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
	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "figura", Kind: KindSticker, Response: fileID}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	if len(fake.SentMessages) != 1 || fake.SentMessages[0].Message.GetStickerMessage() == nil {
		t.Errorf("expected a sticker reply, got %v", fake.SentMessages)
	}
}

func encodeTestOpus(t *testing.T) []byte {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}
	out := filepath.Join(t.TempDir(), "a.ogg")
	if b, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=2",
		"-c:a", "libopus", "-ac", "1", "-f", "ogg", out).CombinedOutput(); err != nil {
		t.Fatalf("encode opus: %v: %s", err, b)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func seedAudioStickerFilter(t *testing.T, m *messages.Message, db *database.DatabaseSingleton, withAudio bool) {
	t.Helper()
	fileID := storage.RandomFilename("webp")
	if err := storage.NewFile(fileID).Write(m.Ctx, encodeTestWebP(t)); err != nil {
		t.Fatalf("seed webp error = %v", err)
	}
	if withAudio {
		if err := storage.NewFile(AudioFilename(fileID)).Write(m.Ctx, encodeTestOpus(t)); err != nil {
			t.Fatalf("seed audio error = %v", err)
		}
	}
	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "figura", Kind: KindAudioSticker, Response: fileID}); err != nil {
		t.Fatalf("seed error = %v", err)
	}
}

func TestAudioFilename(t *testing.T) {
	if got := AudioFilename("abc.webp"); got != "abc.ogg" {
		t.Errorf("AudioFilename = %q, want abc.ogg", got)
	}
}

func TestGetChatFiltersAudioStickerMatch(t *testing.T) {
	db := newTestDB(t)
	t.Chdir(t.TempDir())
	fake := &botfakes.FakeClient{
		UploadFunc: func(ctx context.Context, data []byte, mt whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
			return whatsmeow.UploadResponse{URL: "u", DirectPath: "/d", MediaKey: make([]byte, 32), FileLength: uint64(len(data))}, nil
		},
	}
	m := newTestMessage(t, fake, db, "manda a figura")
	seedAudioStickerFilter(t, m, db, true)

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message, got %d", len(fake.SentMessages))
	}
	audio := fake.SentMessages[0].Message.GetAudioStickerMessage().GetMessage().GetStickerMessage().GetAudioMessage()
	if audio == nil {
		t.Fatalf("expected an audio sticker reply, got %v", fake.SentMessages[0].Message)
	}
	if audio.GetSeconds() != 2 {
		t.Errorf("Seconds = %d, want 2", audio.GetSeconds())
	}
}

func TestGetChatFiltersAudioStickerMissingAudio(t *testing.T) {
	db := newTestDB(t)
	t.Chdir(t.TempDir())
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db, "manda a figura")
	seedAudioStickerFilter(t, m, db, false)

	if err := GetChatFilters(m); err != nil {
		t.Fatalf("GetChatFilters() error = %v", err)
	}
	// Fail reaction + missing-file reply, no sticker.
	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages, got %d", len(fake.SentMessages))
	}
}
