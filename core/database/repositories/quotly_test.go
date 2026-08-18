package repositories_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

func TestQuotlyRepositoryCreateAndFindAllByChat(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file2.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}

	got, err := quotlyRepo.FindAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("FindAllByChat() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 quotly files, got %d: %+v", len(got), got)
	}
}

func TestQuotlyRepositoryFindRandomByChat(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}

	got, err := quotlyRepo.FindRandomByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("FindRandomByChat() error = %v", err)
	}
	if got == nil || got.FileId != "file1.png" {
		t.Errorf("FindRandomByChat() = %+v, want FileId file1.png", got)
	}
}

func TestQuotlyRepositoryFindRandomByChatNoRows(t *testing.T) {
	quotlyRepo := repositories.NewQuotlyRepository(newTestDB(t))

	_, err := quotlyRepo.FindRandomByChat("empty@g.us")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("FindRandomByChat() error = %v, want sql.ErrNoRows (callers rely on this to detect an empty chat)", err)
	}
}

func TestQuotlyRepositoryDelete(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}

	if err := quotlyRepo.Delete("chat1@g.us", "file1.png"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := quotlyRepo.FindAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("FindAllByChat() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected quotly file to be gone after Delete(), got %+v", got)
	}
}

func TestQuotlyRepositoryCreateSentMessageAndFind(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}
	if err := quotlyRepo.CreateSentMessage(&models.QuotlyMessage{
		StanzaID:  "stanza1",
		ChatID:    "chat1@g.us",
		FileId:    "file1.png",
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateSentMessage() error = %v", err)
	}

	got, err := quotlyRepo.FindSentMessageByStanzaID("chat1@g.us", "stanza1")
	if err != nil {
		t.Fatalf("FindSentMessageByStanzaID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a sent message, got nil")
	}
	if got.FileId != "file1.png" {
		t.Errorf("FileId = %q, want %q", got.FileId, "file1.png")
	}
}

func TestQuotlyRepositoryFindSentMessageByStanzaIDNotFound(t *testing.T) {
	quotlyRepo := repositories.NewQuotlyRepository(newTestDB(t))

	got, err := quotlyRepo.FindSentMessageByStanzaID("chat1@g.us", "does-not-exist")
	if err != nil {
		t.Fatalf("FindSentMessageByStanzaID() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for an unknown stanza ID, got %+v", got)
	}
}

func TestQuotlyRepositoryCreateSentMessageCapsAtFivePerFile(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}

	base := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	var stanzaIDs []string
	for i := range 7 {
		id := "stanza" + string(rune('0'+i))
		stanzaIDs = append(stanzaIDs, id)
		if err := quotlyRepo.CreateSentMessage(&models.QuotlyMessage{
			StanzaID:  id,
			ChatID:    "chat1@g.us",
			FileId:    "file1.png",
			CreatedAt: base.Add(time.Duration(i) * time.Second),
		}); err != nil {
			t.Fatalf("CreateSentMessage(%s) error = %v", id, err)
		}
	}

	// The oldest two (stanza0, stanza1) should have aged out of the cap.
	for _, id := range stanzaIDs[:2] {
		got, err := quotlyRepo.FindSentMessageByStanzaID("chat1@g.us", id)
		if err != nil {
			t.Fatalf("FindSentMessageByStanzaID(%s) error = %v", id, err)
		}
		if got != nil {
			t.Errorf("expected %s to have aged out of the cap, but it's still present", id)
		}
	}

	for _, id := range stanzaIDs[2:] {
		got, err := quotlyRepo.FindSentMessageByStanzaID("chat1@g.us", id)
		if err != nil {
			t.Fatalf("FindSentMessageByStanzaID(%s) error = %v", id, err)
		}
		if got == nil {
			t.Errorf("expected %s to still be present within the cap of 5", id)
		}
	}
}

func TestQuotlyRepositoryDeleteCascadesSentMessages(t *testing.T) {
	db := newTestDB(t)
	chatRepo := repositories.NewChatRepository(db)
	quotlyRepo := repositories.NewQuotlyRepository(db)

	if err := chatRepo.Create("chat1@g.us"); err != nil {
		t.Fatalf("Create(chat) error = %v", err)
	}
	if err := quotlyRepo.Create(&models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}); err != nil {
		t.Fatalf("Create(quotly) error = %v", err)
	}
	if err := quotlyRepo.CreateSentMessage(&models.QuotlyMessage{
		StanzaID:  "stanza1",
		ChatID:    "chat1@g.us",
		FileId:    "file1.png",
		CreatedAt: time.Now(),
	}); err != nil {
		t.Fatalf("CreateSentMessage() error = %v", err)
	}

	if err := quotlyRepo.Delete("chat1@g.us", "file1.png"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	got, err := quotlyRepo.FindSentMessageByStanzaID("chat1@g.us", "stanza1")
	if err != nil {
		t.Fatalf("FindSentMessageByStanzaID() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected the sent-message mapping to be cascade-deleted with its quotly file, got %+v", got)
	}
}
