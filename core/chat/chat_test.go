package chat

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
	t.Chdir(filepath.Join(filepath.Dir(file), "..", ".."))

	if err := utils.MigrateSqlite(sqlDB); err != nil {
		t.Fatalf("MigrateSqlite() error = %v", err)
	}

	return &database.DatabaseSingleton{
		Chat:   services.NewChatService(repositories.NewChatRepository(sqlDB)),
		Member: services.NewMemberService(repositories.NewMemberRepository(sqlDB)),
		Filter: services.NewFilterRepository(repositories.NewFilterRepository(sqlDB)),
		Quotly: services.NewQuotlyService(repositories.NewQuotlyRepository(sqlDB)),
	}
}

func newTestMessage(t *testing.T, fake *botfakes.FakeClient, db *database.DatabaseSingleton, chat *models.Chat, text string) *messages.Message {
	t.Helper()
	chatJID := types.NewJID("123456", types.GroupServer)
	chat.ChatID = chatJID.String()
	if _, err := db.Chat.GetOrCreateChat(chatJID.String()); err != nil {
		t.Fatalf("seed chat error = %v", err)
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
		Author: &models.Member{ChatID: chatJID.String(), JID: "111@lid"},
	}
}

func TestChatHandlerStopsAtMutedMember(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	chat := &models.Chat{AllowQuote: 0}
	m := newTestMessage(t, fake, db, chat, "seu babaca")
	m.Author.Silenced = 1

	ChatHandler(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (the revoke), got %d", len(fake.SentMessages))
	}
}

func TestChatHandlerStopsAtProfanity(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	chat := &models.Chat{ProfanityFilterEnabled: 1, WarnBanThreshold: 4}
	m := newTestMessage(t, fake, db, chat, "seu babaca chato")

	ChatHandler(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (revoke + report), got %d", len(fake.SentMessages))
	}
}

func TestChatHandlerStopsAtOffensiveMention(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	chat := &models.Chat{AllowOffensiveReplies: 1, ProfanityFilterEnabled: 0}
	m := newTestMessage(t, fake, db, chat, "bot inutil")

	ChatHandler(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (the comeback reply), got %d", len(fake.SentMessages))
	}
}

func TestChatHandlerFallsThroughToQuotlyDrop(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	chat := &models.Chat{AllowQuote: 0}
	m := newTestMessage(t, fake, db, chat, "just a normal message")

	ChatHandler(m)

	if len(fake.SentMessages) != 0 {
		t.Errorf("expected no messages when nothing matches and quoting is disabled, got %v", fake.SentMessages)
	}
}
