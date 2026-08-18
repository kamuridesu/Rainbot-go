package repositories_test

import (
	"testing"

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
