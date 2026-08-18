package services

import (
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestQuotlyServiceSaveQuotly(t *testing.T) {
	var got *models.QuotlyFile
	repo := &fakeQuotlyRepo{createFn: func(q *models.QuotlyFile) error { got = q; return nil }}
	svc := NewQuotlyService(repo)

	quotly := &models.QuotlyFile{ChatID: "chat1@g.us", FileId: "file1.png"}
	if err := svc.SaveQuotly(quotly); err != nil {
		t.Fatalf("SaveQuotly() error = %v", err)
	}
	if got != quotly {
		t.Errorf("expected repo.Create() to receive the same pointer, got %+v", got)
	}
}

func TestQuotlyServiceGetAllByChat(t *testing.T) {
	want := []*models.QuotlyFile{{FileId: "file1.png"}, {FileId: "file2.png"}}
	repo := &fakeQuotlyRepo{findAllByChatFn: func(chatJid string) ([]*models.QuotlyFile, error) { return want, nil }}
	svc := NewQuotlyService(repo)

	got, err := svc.GetAllByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("GetAllByChat() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetAllByChat() = %v, want 2 files", got)
	}
}

func TestQuotlyServiceGetRandomByChat(t *testing.T) {
	wantErr := errors.New("no rows")
	repo := &fakeQuotlyRepo{findRandomByChatFn: func(chatJid string) (*models.QuotlyFile, error) { return nil, wantErr }}
	svc := NewQuotlyService(repo)

	_, err := svc.GetRandomByChat("chat1@g.us")
	if !errors.Is(err, wantErr) {
		t.Errorf("GetRandomByChat() error = %v, want %v", err, wantErr)
	}
}

func TestQuotlyServiceDeleteQuotly(t *testing.T) {
	var gotChat, gotFile string
	repo := &fakeQuotlyRepo{deleteFn: func(chatJid, fileId string) error {
		gotChat, gotFile = chatJid, fileId
		return nil
	}}
	svc := NewQuotlyService(repo)

	if err := svc.DeleteQuotly("chat1@g.us", "file1.png"); err != nil {
		t.Fatalf("DeleteQuotly() error = %v", err)
	}
	if gotChat != "chat1@g.us" || gotFile != "file1.png" {
		t.Errorf("DeleteQuotly() called repo with (%q, %q), want (%q, %q)", gotChat, gotFile, "chat1@g.us", "file1.png")
	}
}

func TestQuotlyServiceSaveSentMessage(t *testing.T) {
	var got *models.QuotlyMessage
	repo := &fakeQuotlyRepo{createSentMessageFn: func(msg *models.QuotlyMessage) error { got = msg; return nil }}
	svc := NewQuotlyService(repo)

	msg := &models.QuotlyMessage{StanzaID: "s1", ChatID: "chat1@g.us", FileId: "file1.png"}
	if err := svc.SaveSentMessage(msg); err != nil {
		t.Fatalf("SaveSentMessage() error = %v", err)
	}
	if got != msg {
		t.Errorf("expected repo.CreateSentMessage() to receive the same pointer, got %+v", got)
	}
}

func TestQuotlyServiceGetSentMessageByStanzaID(t *testing.T) {
	want := &models.QuotlyMessage{StanzaID: "s1"}
	repo := &fakeQuotlyRepo{
		findSentMessageByStanzaIDFn: func(chatJid, stanzaId string) (*models.QuotlyMessage, error) { return want, nil },
	}
	svc := NewQuotlyService(repo)

	got, err := svc.GetSentMessageByStanzaID("chat1@g.us", "s1")
	if err != nil {
		t.Fatalf("GetSentMessageByStanzaID() error = %v", err)
	}
	if got != want {
		t.Errorf("GetSentMessageByStanzaID() = %+v, want %+v", got, want)
	}
}

func TestQuotlyServiceClose(t *testing.T) {
	svc := NewQuotlyService(&fakeQuotlyRepo{})
	if err := svc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
