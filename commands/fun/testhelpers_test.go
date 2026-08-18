package fun

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

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func newTestSQLiteDB(t *testing.T) *providers.Database {
	t.Helper()

	dbfile := filepath.Join(t.TempDir(), "test.db")
	db, err := providers.InitDB("sqlite3", "file:"+dbfile+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	t.Chdir(repoRoot(t))

	if err := utils.MigrateSqlite(db); err != nil {
		t.Fatalf("MigrateSqlite() error = %v", err)
	}

	return db
}

func newTestDB(t *testing.T) *database.DatabaseSingleton {
	t.Helper()
	sqlDB := newTestSQLiteDB(t)

	return &database.DatabaseSingleton{
		Chat:    services.NewChatService(repositories.NewChatRepository(sqlDB)),
		Member:  services.NewMemberService(repositories.NewMemberRepository(sqlDB)),
		Filter:  services.NewFilterRepository(repositories.NewFilterRepository(sqlDB)),
		Message: services.NewMessageService(repositories.NewMessageRepository(sqlDB)),
		Quotly:  services.NewQuotlyService(repositories.NewQuotlyRepository(sqlDB)),
	}
}

func newTestMessage(t *testing.T, fake *botfakes.FakeClient, db *database.DatabaseSingleton) *messages.Message {
	t.Helper()

	chatJID := types.NewJID("123456", types.GroupServer)
	senderJID := types.NewJID("111", types.HiddenUserServer)

	if _, err := db.Chat.GetOrCreateChat(chatJID.String()); err != nil {
		t.Fatalf("failed to seed chat: %v", err)
	}
	chat, err := db.Chat.Get(chatJID.String())
	if err != nil || chat == nil {
		t.Fatalf("failed to fetch seeded chat: %v", err)
	}

	args := []string{}
	text := ""
	botName := "Rainbot"

	return &messages.Message{
		Ctx:  context.Background(),
		Bot:  &bot.Bot{Client: fake, DB: db, Name: &botName},
		Args: &args,
		Text: &text,
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{Chat: chatJID, Sender: senderJID, IsGroup: true},
				ID:            "orig-stanza-id",
			},
		},
		Chat:   chat,
		Author: &models.Member{ChatID: chatJID.String(), JID: senderJID.String()},
	}
}
