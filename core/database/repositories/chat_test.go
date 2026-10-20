package repositories_test

import (
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

func TestChatRepositoryCreateAndFindById(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))

	if err := repo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	chat, err := repo.FindById("chat1@g.us")
	if err != nil {
		t.Fatalf("FindById() error = %v", err)
	}
	if chat == nil {
		t.Fatal("expected a chat, got nil")
	}

	if chat.ChatID != "chat1@g.us" {
		t.Errorf("ChatID = %q, want %q", chat.ChatID, "chat1@g.us")
	}
	if chat.IsBotEnabled != 1 {
		t.Errorf("IsBotEnabled = %d, want 1", chat.IsBotEnabled)
	}
	if chat.Prefix != "/" {
		t.Errorf("Prefix = %q, want %q", chat.Prefix, "/")
	}
	if chat.WarnBanThreshold != 4 {
		t.Errorf("WarnBanThreshold = %d, want 4", chat.WarnBanThreshold)
	}
	if chat.AllowQuote != 1 {
		t.Errorf("AllowQuote = %d, want 1", chat.AllowQuote)
	}
	if chat.QuoteNMessages != 300 {
		t.Errorf("QuoteNMessages = %d, want 300", chat.QuoteNMessages)
	}
	if chat.AllowOffensiveReplies != 1 {
		t.Errorf("AllowOffensiveReplies = %d, want 1", chat.AllowOffensiveReplies)
	}
	if chat.AllowLLM != 1 {
		t.Errorf("AllowLLM = %d, want 1", chat.AllowLLM)
	}
}

func TestChatRepositoryFindByIdPropagatesNonNoRowsErrors(t *testing.T) {
	db := newTestDB(t)
	repo := repositories.NewChatRepository(db)

	if err := repo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := db.DB.Close(); err != nil {
		t.Fatalf("failed to close db to force an error: %v", err)
	}

	chat, err := repo.FindById("chat1@g.us")
	if err == nil {
		t.Fatal("expected an error when the query fails, got nil")
	}
	if chat != nil {
		t.Errorf("expected nil chat alongside the error, got %+v", chat)
	}
}

func TestChatRepositoryFindByIdNotFound(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))

	chat, err := repo.FindById("does-not-exist@g.us")
	if err != nil {
		t.Fatalf("FindById() error = %v", err)
	}
	if chat != nil {
		t.Errorf("expected nil for a chat that doesn't exist, got %+v", chat)
	}
}

func TestChatRepositoryUpdate(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))

	if err := repo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated := &models.Chat{
		ChatID:                 "chat1@g.us",
		IsBotEnabled:           0,
		Prefix:                 "!",
		AdminOnly:              1,
		CustomProfanityWords:   "foo,bar",
		ProfanityFilterEnabled: 1,
		WarnBanThreshold:       7,
		AllowAdults:            1,
		AllowGames:             0,
		AllowFun:               0,
		WelcomeMessage:         "hi there",
		CountMessages:          0,
		AllowQuote:             0,
		QuoteNMessages:         50,
		AllowOffensiveReplies:  0,
		AllowLLM:               0,
		BotStyle:               "seja formal",
	}

	if err := repo.Update(updated); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := repo.FindById("chat1@g.us")
	if err != nil {
		t.Fatalf("FindById() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a chat, got nil")
	}
	if *got != *updated {
		t.Errorf("FindById() after Update() = %+v, want %+v", *got, *updated)
	}
}

func TestChatRepositoryDelete(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))

	if err := repo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.Delete("chat1@g.us"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := repo.FindById("chat1@g.us")
	if err != nil {
		t.Fatalf("FindById() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected chat to be gone after Delete(), got %+v", got)
	}
}

func TestChatRepositoryMarkLeftKeepsFirstTimestamp(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))
	if err := repo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	first := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := repo.MarkLeft("chat1@g.us", first); err != nil {
		t.Fatalf("MarkLeft() error = %v", err)
	}
	if err := repo.MarkLeft("chat1@g.us", first.Add(48*time.Hour)); err != nil {
		t.Fatalf("second MarkLeft() error = %v", err)
	}

	chat, err := repo.FindById("chat1@g.us")
	if err != nil {
		t.Fatalf("FindById() error = %v", err)
	}
	if chat.LeftAt == nil || !chat.LeftAt.Equal(first) {
		t.Errorf("LeftAt = %v, want %v", chat.LeftAt, first)
	}

	if err := repo.ClearLeft("chat1@g.us"); err != nil {
		t.Fatalf("ClearLeft() error = %v", err)
	}
	chat, _ = repo.FindById("chat1@g.us")
	if chat.LeftAt != nil {
		t.Errorf("LeftAt after ClearLeft = %v, want nil", chat.LeftAt)
	}
}

func TestChatRepositoryFindLeftBefore(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	for _, id := range []string{"old@g.us", "recent@g.us", "active@g.us"} {
		if err := repo.Create(id); err != nil {
			t.Fatalf("Create(%s) error = %v", id, err)
		}
	}
	repo.MarkLeft("old@g.us", now.Add(-8*24*time.Hour))
	repo.MarkLeft("recent@g.us", now.Add(-2*24*time.Hour))

	got, err := repo.FindLeftBefore(now.Add(-7 * 24 * time.Hour))
	if err != nil {
		t.Fatalf("FindLeftBefore() error = %v", err)
	}
	if len(got) != 1 || got[0] != "old@g.us" {
		t.Errorf("FindLeftBefore() = %v, want [old@g.us]", got)
	}
}

func TestChatRepositoryPurgeDeletesAllChatData(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)
	filterRepo := repositories.NewFilterRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	for _, id := range []string{"gone@g.us", "kept@g.us"} {
		if err := chatRepo.Create(id); err != nil {
			t.Fatalf("Create(%s) error = %v", id, err)
		}
		if err := memberRepo.Create(id, "123@lid"); err != nil {
			t.Fatalf("Create(member) error = %v", err)
		}
		if err := filterRepo.Create(&models.Filter{ChatID: id, Pattern: "hi", Kind: "image", Response: id + ".jpg"}); err != nil {
			t.Fatalf("Create(filter) error = %v", err)
		}
		if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: id, FileId: id + ".png"}); err != nil {
			t.Fatalf("Create(quotly) error = %v", err)
		}
		if err := quotlyRepo.CreateSentMessage(&models.QuotlyMessage{StanzaID: id + "-q", ChatID: id, FileId: id + ".png", CreatedAt: now}); err != nil {
			t.Fatalf("CreateSentMessage() error = %v", err)
		}
		if err := messageRepo.Create(&models.Message{StanzaID: id + "-m", ChatID: id, SenderJID: "123@lid", MessageText: "hello", CreatedAt: now}); err != nil {
			t.Fatalf("Create(message) error = %v", err)
		}
	}
	chatRepo.MarkLeft("gone@g.us", now.Add(-8*24*time.Hour))
	cutoff := now.Add(-7 * 24 * time.Hour)

	purged, err := chatRepo.Purge("gone@g.us", cutoff)
	if err != nil {
		t.Fatalf("Purge() error = %v", err)
	}
	if !purged {
		t.Fatal("Purge() = false, want true")
	}

	for _, table := range []string{"chat", "member", "filter", "quotly", "quotly_message", "messages"} {
		var gone, kept int
		db.DB.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE chatId = ?", "gone@g.us").Scan(&gone)
		db.DB.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE chatId = ?", "kept@g.us").Scan(&kept)
		if gone != 0 {
			t.Errorf("%s still has %d rows for purged chat", table, gone)
		}
		if kept != 1 {
			t.Errorf("%s has %d rows for untouched chat, want 1", table, kept)
		}
	}
}

func TestChatRepositoryPurgeSkipsChatsNotPastCutoff(t *testing.T) {
	repo := repositories.NewChatRepository(newTestDB(t))
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	cutoff := now.Add(-7 * 24 * time.Hour)

	repo.Create("active@g.us")
	repo.Create("recent@g.us")
	repo.MarkLeft("recent@g.us", now.Add(-24*time.Hour))

	for _, id := range []string{"active@g.us", "recent@g.us", "missing@g.us"} {
		purged, err := repo.Purge(id, cutoff)
		if err != nil {
			t.Fatalf("Purge(%s) error = %v", id, err)
		}
		if purged {
			t.Errorf("Purge(%s) = true, want false", id)
		}
	}
	if chat, _ := repo.FindById("active@g.us"); chat == nil {
		t.Error("active chat was deleted")
	}
	if chat, _ := repo.FindById("recent@g.us"); chat == nil {
		t.Error("recently left chat was deleted")
	}
}
