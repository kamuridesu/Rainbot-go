package services

import (
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestChatServiceGetOrCreateChatExisting(t *testing.T) {
	existing := &models.Chat{ChatID: "chat1@g.us", Prefix: "!"}
	repo := &fakeChatRepo{
		findByIdFn: func(jid string) (*models.Chat, error) { return existing, nil },
	}
	svc := NewChatService(repo)

	got, err := svc.GetOrCreateChat("chat1@g.us")
	if err != nil {
		t.Fatalf("GetOrCreateChat() error = %v", err)
	}
	if got != existing {
		t.Errorf("expected the existing chat to be returned as-is, got %+v", got)
	}
	if len(repo.createCalls) != 0 {
		t.Error("expected Create() not to be called for an existing chat")
	}
}

func TestChatServiceGetOrCreateChatNew(t *testing.T) {
	repo := &fakeChatRepo{
		findByIdFn: func(jid string) (*models.Chat, error) { return nil, nil },
	}
	svc := NewChatService(repo)

	got, err := svc.GetOrCreateChat("chat1@g.us")
	if err != nil {
		t.Fatalf("GetOrCreateChat() error = %v", err)
	}
	if len(repo.createCalls) != 1 || repo.createCalls[0] != "chat1@g.us" {
		t.Errorf("expected Create() to be called once with chat1@g.us, got %v", repo.createCalls)
	}

	want := &models.Chat{
		ChatID:                 "chat1@g.us",
		IsBotEnabled:           1,
		Prefix:                 "/",
		AdminOnly:              0,
		ProfanityFilterEnabled: 0,
		CustomProfanityWords:   "",
		WarnBanThreshold:       4,
		AllowAdults:            0,
		AllowGames:             1,
		AllowFun:               1,
		WelcomeMessage:         "",
		CountMessages:          1,
		AllowQuote:             1,
		QuoteNMessages:         300,
		AllowOffensiveReplies:  1,
	}
	if *got != *want {
		t.Errorf("GetOrCreateChat() = %+v, want %+v", *got, *want)
	}
}

func TestChatServiceGetOrCreateChatFindError(t *testing.T) {
	wantErr := errors.New("db exploded")
	repo := &fakeChatRepo{
		findByIdFn: func(jid string) (*models.Chat, error) { return nil, wantErr },
	}
	svc := NewChatService(repo)

	_, err := svc.GetOrCreateChat("chat1@g.us")
	if !errors.Is(err, wantErr) {
		t.Errorf("GetOrCreateChat() error = %v, want %v", err, wantErr)
	}
}

func TestChatServiceGetOrCreateChatCreateError(t *testing.T) {
	wantErr := errors.New("insert failed")
	repo := &fakeChatRepo{
		findByIdFn: func(jid string) (*models.Chat, error) { return nil, nil },
		createErr:  wantErr,
	}
	svc := NewChatService(repo)

	_, err := svc.GetOrCreateChat("chat1@g.us")
	if !errors.Is(err, wantErr) {
		t.Errorf("GetOrCreateChat() error = %v, want %v", err, wantErr)
	}
}

func TestChatServiceGet(t *testing.T) {
	existing := &models.Chat{ChatID: "chat1@g.us"}
	repo := &fakeChatRepo{
		findByIdFn: func(jid string) (*models.Chat, error) { return existing, nil },
	}
	svc := NewChatService(repo)

	got, err := svc.Get("chat1@g.us")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got != existing {
		t.Errorf("Get() = %+v, want %+v", got, existing)
	}
}

func TestChatServiceUpdateChatRejectsInvalidPrefixLength(t *testing.T) {
	repo := &fakeChatRepo{}
	svc := NewChatService(repo)

	err := svc.UpdateChat(&models.Chat{ChatID: "chat1@g.us", Prefix: "!!"})
	if err == nil {
		t.Fatal("expected an error for a multi-character prefix")
	}
	if len(repo.updateCalls) != 0 {
		t.Error("expected Update() not to be called when validation fails")
	}
}

func TestChatServiceUpdateChatRejectsEmptyPrefix(t *testing.T) {
	repo := &fakeChatRepo{}
	svc := NewChatService(repo)

	if err := svc.UpdateChat(&models.Chat{ChatID: "chat1@g.us", Prefix: ""}); err == nil {
		t.Fatal("expected an error for an empty prefix")
	}
}

func TestChatServiceUpdateChatValid(t *testing.T) {
	repo := &fakeChatRepo{}
	svc := NewChatService(repo)

	chat := &models.Chat{ChatID: "chat1@g.us", Prefix: "!"}
	if err := svc.UpdateChat(chat); err != nil {
		t.Fatalf("UpdateChat() error = %v", err)
	}
	if len(repo.updateCalls) != 1 || repo.updateCalls[0] != chat {
		t.Errorf("expected Update() to be called once with the chat, got %v", repo.updateCalls)
	}
}

func TestChatServiceUpdateChatPropagatesRepoError(t *testing.T) {
	wantErr := errors.New("update failed")
	repo := &fakeChatRepo{updateFn: func(chat *models.Chat) error { return wantErr }}
	svc := NewChatService(repo)

	err := svc.UpdateChat(&models.Chat{ChatID: "chat1@g.us", Prefix: "!"})
	if !errors.Is(err, wantErr) {
		t.Errorf("UpdateChat() error = %v, want %v", err, wantErr)
	}
}

func TestChatServiceClose(t *testing.T) {
	svc := NewChatService(&fakeChatRepo{})
	if err := svc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
