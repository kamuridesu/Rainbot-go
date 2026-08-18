package repositories_test

import (
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

func TestFilterRepositoryCreateAndFindAllByChat(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	filterRepo := repositories.NewFilterRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}

	filters := []*models.Filter{
		{ChatID: "chat1@g.us", Pattern: "oi", Kind: "text", Response: "ola!"},
		{ChatID: "chat1@g.us", Pattern: "tchau", Kind: "text", Response: "ate mais!"},
	}
	for _, f := range filters {
		if err := filterRepo.Create(f); err != nil {
			t.Fatalf("Create(filter %q) error = %v", f.Pattern, err)
		}
	}

	got, err := filterRepo.FindAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("FindAllByChat() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 filters, got %d: %+v", len(got), got)
	}

	byPattern := map[string]*models.Filter{}
	for _, f := range got {
		byPattern[f.Pattern] = f
	}
	if byPattern["oi"] == nil || byPattern["oi"].Response != "ola!" {
		t.Errorf("unexpected filter for pattern 'oi': %+v", byPattern["oi"])
	}
	if byPattern["tchau"] == nil || byPattern["tchau"].Response != "ate mais!" {
		t.Errorf("unexpected filter for pattern 'tchau': %+v", byPattern["tchau"])
	}
}

func TestFilterRepositoryFindAllByChatEmpty(t *testing.T) {
	filterRepo := repositories.NewFilterRepository(newTestDB(t))

	got, err := filterRepo.FindAllByChat("no-filters@g.us")
	if err != nil {
		t.Fatalf("FindAllByChat() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no filters, got %d: %+v", len(got), got)
	}
}

func TestFilterRepositoryDelete(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	filterRepo := repositories.NewFilterRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := filterRepo.Create(&models.Filter{ChatID: "chat1@g.us", Pattern: "oi", Kind: "text", Response: "ola!"}); err != nil {
		t.Fatalf("Create(filter) error = %v", err)
	}

	if err := filterRepo.Delete("chat1@g.us", "oi"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := filterRepo.FindAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("FindAllByChat() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected filter to be gone after Delete(), got %+v", got)
	}
}

func TestFilterRepositoryDeleteOnlyMatchesExactPattern(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	filterRepo := repositories.NewFilterRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := filterRepo.Create(&models.Filter{ChatID: "chat1@g.us", Pattern: "oi", Kind: "text", Response: "ola!"}); err != nil {
		t.Fatalf("Create(filter) error = %v", err)
	}

	if err := filterRepo.Delete("chat1@g.us", "nao-existe"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := filterRepo.FindAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("FindAllByChat() error = %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected the unrelated filter to survive, got %d filters: %+v", len(got), got)
	}
}
