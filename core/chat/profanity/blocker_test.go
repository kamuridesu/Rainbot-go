package profanity

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
		Member: services.NewMemberService(repositories.NewMemberRepository(sqlDB)),
	}
}

func newTestMessage(t *testing.T, fake *botfakes.FakeClient, chat *models.Chat, text string) *messages.Message {
	t.Helper()
	return &messages.Message{
		Ctx:  context.Background(),
		Bot:  &bot.Bot{Client: fake, DB: newTestDB(t)},
		Text: &text,
		Chat: chat,
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Chat:   types.NewJID("123", types.GroupServer),
					Sender: types.NewJID("111", types.HiddenUserServer),
				},
				ID: "stanza-1",
			},
		},
		Author: &models.Member{JID: "111@lid", Warns: 0},
	}
}

func TestCheckForWordDisabledByChatSetting(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, &models.Chat{ProfanityFilterEnabled: 0}, "seu babaca")

	if CheckForWord(m) {
		t.Error("expected false when the profanity filter is disabled")
	}
	if len(fake.SentMessages) != 0 {
		t.Error("expected no messages sent")
	}
}

func TestCheckForWordCleanMessage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, &models.Chat{ProfanityFilterEnabled: 1}, "bom dia pessoal")

	if CheckForWord(m) {
		t.Error("expected false for a clean message")
	}
}

func TestCheckForWordBlocksObsceneWord(t *testing.T) {
	fake := &botfakes.FakeClient{}
	chat := &models.Chat{ProfanityFilterEnabled: 1, WarnBanThreshold: 4}
	m := newTestMessage(t, fake, chat, "seu babaca chato")

	if !CheckForWord(m) {
		t.Fatal("expected true for a message containing an obscene word")
	}
	if m.Author.Warns != 1 {
		t.Errorf("expected the author to receive a warn, got %d", m.Author.Warns)
	}
	// One revoke (delete the message) and one report message with mentions.
	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (revoke + report), got %d", len(fake.SentMessages))
	}
}

func TestCheckForWordBlocksCustomWord(t *testing.T) {
	fake := &botfakes.FakeClient{}
	chat := &models.Chat{ProfanityFilterEnabled: 1, WarnBanThreshold: 4, CustomProfanityWords: "banana"}
	m := newTestMessage(t, fake, chat, "eu quero uma banana")

	if !CheckForWord(m) {
		t.Fatal("expected true for a message containing a custom-blocked word")
	}
}
