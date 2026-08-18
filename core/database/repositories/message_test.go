package repositories_test

import (
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/database/repositories"
)

func TestMessageRepositoryCreateAndFindByStanzaID(t *testing.T) {
	repo := repositories.NewMessageRepository(newTestDB(t))

	quoted := "quoted-stanza"
	msg := &models.Message{
		StanzaID:       "stanza1",
		ChatID:         "chat1@g.us",
		SenderJID:      "111@lid",
		MessageText:    "hello",
		QuotedStanzaID: &quoted,
		CreatedAt:      time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC),
	}
	if err := repo.Create(msg); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByStanzaID("stanza1")
	if err != nil {
		t.Fatalf("FindByStanzaID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a message, got nil")
	}
	if got.MessageText != "hello" || got.SenderJID != "111@lid" || got.ChatID != "chat1@g.us" {
		t.Errorf("unexpected message: %+v", got)
	}
	if got.QuotedStanzaID == nil || *got.QuotedStanzaID != "quoted-stanza" {
		t.Errorf("QuotedStanzaID = %v, want %q", got.QuotedStanzaID, "quoted-stanza")
	}
}

func TestMessageRepositoryCreateWithoutQuotedStanza(t *testing.T) {
	repo := repositories.NewMessageRepository(newTestDB(t))

	msg := &models.Message{
		StanzaID:    "stanza1",
		ChatID:      "chat1@g.us",
		SenderJID:   "111@lid",
		MessageText: "no reply here",
		CreatedAt:   time.Now(),
	}
	if err := repo.Create(msg); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.FindByStanzaID("stanza1")
	if err != nil {
		t.Fatalf("FindByStanzaID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a message, got nil")
	}
	if got.QuotedStanzaID != nil {
		t.Errorf("QuotedStanzaID = %v, want nil", got.QuotedStanzaID)
	}
}

func TestMessageRepositoryFindByStanzaIDNotFound(t *testing.T) {
	repo := repositories.NewMessageRepository(newTestDB(t))

	got, err := repo.FindByStanzaID("does-not-exist")
	if err != nil {
		t.Fatalf("FindByStanzaID() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for a missing message, got %+v", got)
	}
}

func TestMessageRepositoryCreateUpsertsOnEdit(t *testing.T) {
	repo := repositories.NewMessageRepository(newTestDB(t))

	ts := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	original := &models.Message{
		StanzaID:    "stanza1",
		ChatID:      "chat1@g.us",
		SenderJID:   "111@lid",
		MessageText: "original text",
		CreatedAt:   ts,
	}
	if err := repo.Create(original); err != nil {
		t.Fatalf("initial Create() error = %v", err)
	}

	edited := &models.Message{
		StanzaID:    "stanza1",
		ChatID:      "chat1@g.us",
		SenderJID:   "111@lid",
		MessageText: "edited text",
		CreatedAt:   ts,
	}
	if err := repo.Create(edited); err != nil {
		t.Fatalf("edit Create() error = %v", err)
	}

	got, err := repo.FindByStanzaID("stanza1")
	if err != nil {
		t.Fatalf("FindByStanzaID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected a message, got nil")
	}
	if got.MessageText != "edited text" {
		t.Errorf("MessageText = %q, want %q (edit should overwrite, not duplicate)", got.MessageText, "edited text")
	}
}

func TestMessageRepositoryFindMessagesAfter(t *testing.T) {
	repo := repositories.NewMessageRepository(newTestDB(t))

	base := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	messages := []*models.Message{
		{StanzaID: "s1", ChatID: "chat1@g.us", SenderJID: "111@lid", MessageText: "first", CreatedAt: base},
		{StanzaID: "s2", ChatID: "chat1@g.us", SenderJID: "111@lid", MessageText: "second", CreatedAt: base.Add(1 * time.Minute)},
		{StanzaID: "s3", ChatID: "chat1@g.us", SenderJID: "111@lid", MessageText: "third", CreatedAt: base.Add(2 * time.Minute)},
		// Before the "since" cutoff, must be excluded.
		{StanzaID: "s0", ChatID: "chat1@g.us", SenderJID: "111@lid", MessageText: "too old", CreatedAt: base.Add(-1 * time.Minute)},
		// Different chat, must be excluded regardless of time.
		{StanzaID: "s4", ChatID: "chat2@g.us", SenderJID: "222@lid", MessageText: "other chat", CreatedAt: base.Add(1 * time.Minute)},
	}
	for _, m := range messages {
		if err := repo.Create(m); err != nil {
			t.Fatalf("Create(%s) error = %v", m.StanzaID, err)
		}
	}

	got, err := repo.FindMessagesAfter("chat1@g.us", base, 10)
	if err != nil {
		t.Fatalf("FindMessagesAfter() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 messages, got %d: %+v", len(got), got)
	}

	wantOrder := []string{"s1", "s2", "s3"}
	for i, want := range wantOrder {
		if got[i].StanzaID != want {
			t.Errorf("got[%d].StanzaID = %q, want %q (expected ascending createdAt order)", i, got[i].StanzaID, want)
		}
	}
}

func TestMessageRepositoryFindMessagesAfterRespectsLimit(t *testing.T) {
	repo := repositories.NewMessageRepository(newTestDB(t))

	base := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	for i := range 5 {
		m := &models.Message{
			StanzaID:    "s" + string(rune('0'+i)),
			ChatID:      "chat1@g.us",
			SenderJID:   "111@lid",
			MessageText: "msg",
			CreatedAt:   base.Add(time.Duration(i) * time.Minute),
		}
		if err := repo.Create(m); err != nil {
			t.Fatalf("Create(%s) error = %v", m.StanzaID, err)
		}
	}

	got, err := repo.FindMessagesAfter("chat1@g.us", base, 2)
	if err != nil {
		t.Fatalf("FindMessagesAfter() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected limit of 2 messages, got %d: %+v", len(got), got)
	}
}
