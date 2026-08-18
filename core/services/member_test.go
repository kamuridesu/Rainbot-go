package services

import (
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestMemberServiceGetOrCreateMemberExisting(t *testing.T) {
	existing := &models.Member{ChatID: "chat1@g.us", JID: "111@lid"}
	repo := &fakeMemberRepo{
		findByChatAndIdFn: func(chatJid, memberJid string) (*models.Member, error) { return existing, nil },
	}
	svc := NewMemberService(repo)

	got, err := svc.GetOrCreateMember("chat1@g.us", "111@lid")
	if err != nil {
		t.Fatalf("GetOrCreateMember() error = %v", err)
	}
	if got != existing {
		t.Errorf("expected the existing member to be returned as-is, got %+v", got)
	}
	if len(repo.createCalls) != 0 {
		t.Error("expected Create() not to be called for an existing member")
	}
}

func TestMemberServiceGetOrCreateMemberNew(t *testing.T) {
	repo := &fakeMemberRepo{
		findByChatAndIdFn: func(chatJid, memberJid string) (*models.Member, error) { return nil, nil },
	}
	svc := NewMemberService(repo)

	got, err := svc.GetOrCreateMember("chat1@g.us", "111@lid")
	if err != nil {
		t.Fatalf("GetOrCreateMember() error = %v", err)
	}
	if len(repo.createCalls) != 1 || repo.createCalls[0] != [2]string{"chat1@g.us", "111@lid"} {
		t.Errorf("expected Create() to be called once with (chat1@g.us, 111@lid), got %v", repo.createCalls)
	}

	want := &models.Member{ChatID: "chat1@g.us", JID: "111@lid"}
	if *got != *want {
		t.Errorf("GetOrCreateMember() = %+v, want %+v", *got, *want)
	}
}

func TestMemberServiceGetOrCreateMemberRejectsNonLidJID(t *testing.T) {
	repo := &fakeMemberRepo{
		findByChatAndIdFn: func(chatJid, memberJid string) (*models.Member, error) { return nil, nil },
	}
	svc := NewMemberService(repo)

	_, err := svc.GetOrCreateMember("chat1@g.us", "5511999999999@s.whatsapp.net")
	if err == nil {
		t.Fatal("expected an error for a non-@lid JID")
	}
	if len(repo.createCalls) != 0 {
		t.Error("expected Create() not to be called for a rejected JID")
	}
}

func TestMemberServiceGetOrCreateMemberFindError(t *testing.T) {
	wantErr := errors.New("db exploded")
	repo := &fakeMemberRepo{
		findByChatAndIdFn: func(chatJid, memberJid string) (*models.Member, error) { return nil, wantErr },
	}
	svc := NewMemberService(repo)

	_, err := svc.GetOrCreateMember("chat1@g.us", "111@lid")
	if !errors.Is(err, wantErr) {
		t.Errorf("GetOrCreateMember() error = %v, want %v", err, wantErr)
	}
}

func TestMemberServiceGetOrCreateMemberCreateError(t *testing.T) {
	wantErr := errors.New("insert failed")
	repo := &fakeMemberRepo{
		findByChatAndIdFn: func(chatJid, memberJid string) (*models.Member, error) { return nil, nil },
		createErr:         wantErr,
	}
	svc := NewMemberService(repo)

	_, err := svc.GetOrCreateMember("chat1@g.us", "111@lid")
	if !errors.Is(err, wantErr) {
		t.Errorf("GetOrCreateMember() error = %v, want %v", err, wantErr)
	}
}

func TestMemberServiceUpdate(t *testing.T) {
	var got *models.Member
	repo := &fakeMemberRepo{updateFn: func(member *models.Member) error { got = member; return nil }}
	svc := NewMemberService(repo)

	member := &models.Member{ChatID: "chat1@g.us", JID: "111@lid", Warns: 2}
	if err := svc.Update(member); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got != member {
		t.Errorf("expected repo.Update() to receive the same member pointer, got %+v", got)
	}
}

func TestMemberServiceGetByChat(t *testing.T) {
	want := []*models.Member{{JID: "111@lid"}, {JID: "222@lid"}}
	repo := &fakeMemberRepo{getAllByChatFn: func(chatJid string) ([]*models.Member, error) { return want, nil }}
	svc := NewMemberService(repo)

	got, err := svc.GetByChat("chat1@g.us")
	if err != nil {
		t.Fatalf("GetByChat() error = %v", err)
	}
	if len(got) != 2 {
		t.Errorf("GetByChat() = %v, want 2 members", got)
	}
}

func TestMemberServiceClose(t *testing.T) {
	svc := NewMemberService(&fakeMemberRepo{})
	if err := svc.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
