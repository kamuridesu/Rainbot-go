package repositories_test

import (
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

func TestMemberRepositoryCreateAndFind(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := memberRepo.Create("chat1@g.us", "111@lid"); err != nil {
		t.Fatalf("Create(member) error = %v", err)
	}

	got, err := memberRepo.FindByChatAndId("chat1@g.us", "111@lid")
	if err != nil {
		t.Fatalf("FindByChatAndId() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a member, got nil")
	}
	if got.ChatID != "chat1@g.us" || got.JID != "111@lid" {
		t.Errorf("unexpected member: %+v", got)
	}
	if got.Warns != 0 || got.Points != 0 || got.Messages != 0 || got.Silenced != 0 {
		t.Errorf("expected zero-value counters on creation, got %+v", got)
	}
}

func TestMemberRepositoryFindByChatAndIdNotFound(t *testing.T) {
	memberRepo := repositories.NewMemberRepository(newTestDB(t))

	got, err := memberRepo.FindByChatAndId("chat1@g.us", "nope@lid")
	if err != nil {
		t.Fatalf("FindByChatAndId() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for a missing member, got %+v", got)
	}
}

func TestMemberRepositoryFindByChatAndIdRejectsOldFormatJID(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := memberRepo.Create("chat1@g.us", "5511999999999@s.whatsapp.net"); err != nil {
		t.Fatalf("Create(member) error = %v", err)
	}

	_, err := memberRepo.FindByChatAndId("chat1@g.us", "5511999999999@s.whatsapp.net")
	if err == nil {
		t.Fatal("expected an error for an old-format (s.whatsapp.net) JID")
	}
}

func TestMemberRepositoryUpdate(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := memberRepo.Create("chat1@g.us", "111@lid"); err != nil {
		t.Fatalf("Create(member) error = %v", err)
	}

	updated := &models.Member{
		ChatID:   "chat1@g.us",
		JID:      "111@lid",
		Warns:    2,
		Points:   10,
		Messages: 50,
		Silenced: 1,
	}
	if err := memberRepo.Update(updated); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := memberRepo.FindByChatAndId("chat1@g.us", "111@lid")
	if err != nil {
		t.Fatalf("FindByChatAndId() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a member, got nil")
	}
	if *got != *updated {
		t.Errorf("FindByChatAndId() after Update() = %+v, want %+v", *got, *updated)
	}
}

func TestMemberRepositoryGetAllByChatExcludesOldFormatJIDs(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	memberRepo := repositories.NewMemberRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := memberRepo.Create("chat1@g.us", "111@lid"); err != nil {
		t.Fatalf("Create(member) error = %v", err)
	}
	if err := memberRepo.Create("chat1@g.us", "222@lid"); err != nil {
		t.Fatalf("Create(member) error = %v", err)
	}
	if err := memberRepo.Create("chat1@g.us", "5511999999999@s.whatsapp.net"); err != nil {
		t.Fatalf("Create(member) error = %v", err)
	}
	if err := chatRepo.Create("chat2@g.us"); err != nil {
		t.Fatalf("Create(chat2) error = %v", err)
	}
	if err := memberRepo.Create("chat2@g.us", "333@lid"); err != nil {
		t.Fatalf("Create(member in other chat) error = %v", err)
	}

	members, err := memberRepo.GetAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("GetAllByChat() error = %v", err)
	}

	if len(members) != 2 {
		t.Fatalf("expected 2 members (old-format JID excluded), got %d: %+v", len(members), members)
	}

	seen := map[string]bool{}
	for _, m := range members {
		seen[m.JID] = true
	}
	if !seen["111@lid"] || !seen["222@lid"] {
		t.Errorf("expected 111@lid and 222@lid in results, got %+v", members)
	}
}

func TestMemberRepositoryMarkAndClearLeft(t *testing.T) {
	db := newTestDB(t)
	repositories.NewChatRepository(db).Create("chat1@g.us")
	repo := repositories.NewMemberRepository(db)
	if err := repo.Create("chat1@g.us", "123@lid"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	first := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := repo.MarkLeft("chat1@g.us", "123@lid", first); err != nil {
		t.Fatalf("MarkLeft() error = %v", err)
	}
	repo.MarkLeft("chat1@g.us", "123@lid", first.Add(time.Hour))

	member, err := repo.FindByChatAndId("chat1@g.us", "123@lid")
	if err != nil {
		t.Fatalf("FindByChatAndId() error = %v", err)
	}
	if member.LeftAt == nil || !member.LeftAt.Equal(first) {
		t.Errorf("LeftAt = %v, want %v", member.LeftAt, first)
	}

	if err := repo.ClearLeft("chat1@g.us", "123@lid"); err != nil {
		t.Fatalf("ClearLeft() error = %v", err)
	}
	member, _ = repo.FindByChatAndId("chat1@g.us", "123@lid")
	if member.LeftAt != nil {
		t.Errorf("LeftAt after ClearLeft = %v, want nil", member.LeftAt)
	}
}

func TestMemberRepositoryDeleteLeftBefore(t *testing.T) {
	db := newTestDB(t)
	repositories.NewChatRepository(db).Create("chat1@g.us")
	repositories.NewChatRepository(db).Create("chat2@g.us")
	repo := repositories.NewMemberRepository(db)
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)

	repo.Create("chat1@g.us", "old@lid")
	repo.Create("chat1@g.us", "recent@lid")
	repo.Create("chat1@g.us", "active@lid")
	repo.Create("chat2@g.us", "old@lid")
	repo.MarkLeft("chat1@g.us", "old@lid", now.Add(-8*24*time.Hour))
	repo.MarkLeft("chat1@g.us", "recent@lid", now.Add(-24*time.Hour))

	removed, err := repo.DeleteLeftBefore(now.Add(-7 * 24 * time.Hour))
	if err != nil {
		t.Fatalf("DeleteLeftBefore() error = %v", err)
	}
	if removed != 1 {
		t.Errorf("DeleteLeftBefore() = %d, want 1", removed)
	}

	if m, _ := repo.FindByChatAndId("chat1@g.us", "old@lid"); m != nil {
		t.Error("member who left over a week ago should be deleted")
	}
	for _, key := range [][2]string{{"chat1@g.us", "recent@lid"}, {"chat1@g.us", "active@lid"}, {"chat2@g.us", "old@lid"}} {
		if m, _ := repo.FindByChatAndId(key[0], key[1]); m == nil {
			t.Errorf("member %s in %s should be kept", key[1], key[0])
		}
	}
}
