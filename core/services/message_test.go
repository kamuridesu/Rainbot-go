package services

import (
	"errors"
	"testing"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestMessageServiceSaveMessage(t *testing.T) {
	var got *models.Message
	repo := &fakeMessageRepo{createFn: func(msg *models.Message) error { got = msg; return nil }}
	svc := NewMessageService(repo)

	msg := &models.Message{StanzaID: "s1", ChatID: "chat1@g.us"}
	if err := svc.SaveMessage(msg); err != nil {
		t.Fatalf("SaveMessage() error = %v", err)
	}
	if got != msg {
		t.Errorf("expected repo.Create() to receive the same message pointer, got %+v", got)
	}
}

func TestMessageServiceSaveMessageError(t *testing.T) {
	wantErr := errors.New("insert failed")
	repo := &fakeMessageRepo{createFn: func(msg *models.Message) error { return wantErr }}
	svc := NewMessageService(repo)

	if err := svc.SaveMessage(&models.Message{}); !errors.Is(err, wantErr) {
		t.Errorf("SaveMessage() error = %v, want %v", err, wantErr)
	}
}

func TestMessageServiceGetMessage(t *testing.T) {
	want := &models.Message{StanzaID: "s1"}
	repo := &fakeMessageRepo{findByStanzaIDFn: func(stanzaID string) (*models.Message, error) { return want, nil }}
	svc := NewMessageService(repo)

	got, err := svc.GetMessage("s1")
	if err != nil {
		t.Fatalf("GetMessage() error = %v", err)
	}
	if got != want {
		t.Errorf("GetMessage() = %+v, want %+v", got, want)
	}
}

func TestMessageServiceGetMessageRange(t *testing.T) {
	since := time.Now()
	var gotChat string
	var gotSince time.Time
	var gotLimit int
	want := []*models.Message{{StanzaID: "s1"}, {StanzaID: "s2"}}

	repo := &fakeMessageRepo{
		findMessagesAfter: func(chatId string, s time.Time, limit int) ([]*models.Message, error) {
			gotChat, gotSince, gotLimit = chatId, s, limit
			return want, nil
		},
	}
	svc := NewMessageService(repo)

	got, err := svc.GetMessageRange("chat1@g.us", since, 5)
	if err != nil {
		t.Fatalf("GetMessageRange() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetMessageRange() = %v, want 2 messages", got)
	}
	if gotChat != "chat1@g.us" || !gotSince.Equal(since) || gotLimit != 5 {
		t.Errorf("GetMessageRange() called repo with (%q, %v, %d), want (%q, %v, %d)", gotChat, gotSince, gotLimit, "chat1@g.us", since, 5)
	}
}

func TestMessageServiceClose(t *testing.T) {
	svc := NewMessageService(&fakeMessageRepo{})
	if err := svc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
