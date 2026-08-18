package services

import (
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestFilterServiceGetFilters(t *testing.T) {
	want := []*models.Filter{{ChatID: "chat1@g.us", Pattern: "oi"}}
	repo := &fakeFilterRepo{
		findAllByChatFn: func(chatJid string) ([]*models.Filter, error) {
			if chatJid != "chat1@g.us" {
				t.Errorf("FindAllByChat() called with %q, want %q", chatJid, "chat1@g.us")
			}
			return want, nil
		},
	}
	svc := NewFilterRepository(repo)

	got, err := svc.GetFilters("chat1@g.us")
	if err != nil {
		t.Fatalf("GetFilters() error = %v", err)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("GetFilters() = %v, want %v", got, want)
	}
}

func TestFilterServiceGetFiltersError(t *testing.T) {
	wantErr := errors.New("db exploded")
	repo := &fakeFilterRepo{findAllByChatFn: func(chatJid string) ([]*models.Filter, error) { return nil, wantErr }}
	svc := NewFilterRepository(repo)

	_, err := svc.GetFilters("chat1@g.us")
	if !errors.Is(err, wantErr) {
		t.Errorf("GetFilters() error = %v, want %v", err, wantErr)
	}
}

func TestFilterServiceNewFilter(t *testing.T) {
	var got *models.Filter
	repo := &fakeFilterRepo{createFn: func(filter *models.Filter) error { got = filter; return nil }}
	svc := NewFilterRepository(repo)

	filter := &models.Filter{ChatID: "chat1@g.us", Pattern: "oi", Kind: "text", Response: "ola"}
	if err := svc.NewFilter(filter); err != nil {
		t.Fatalf("NewFilter() error = %v", err)
	}
	if got != filter {
		t.Errorf("expected repo.Create() to receive the same filter pointer, got %+v", got)
	}
}

func TestFilterServiceDelete(t *testing.T) {
	var gotChat, gotPattern string
	repo := &fakeFilterRepo{deleteFn: func(chatJid, pattern string) error {
		gotChat, gotPattern = chatJid, pattern
		return nil
	}}
	svc := NewFilterRepository(repo)

	if err := svc.Delete("chat1@g.us", "oi"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if gotChat != "chat1@g.us" || gotPattern != "oi" {
		t.Errorf("Delete() called repo with (%q, %q), want (%q, %q)", gotChat, gotPattern, "chat1@g.us", "oi")
	}
}

func TestFilterServiceClose(t *testing.T) {
	svc := NewFilterRepository(&fakeFilterRepo{})
	if err := svc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
