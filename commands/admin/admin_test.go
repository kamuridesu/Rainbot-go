package admin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func TestBanUserSuccess(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.MentionedMembers = []*models.Member{{JID: "111@lid"}, {JID: "222@lid"}}

	BanUser(m)

	if len(fake.UpdatedGroupParticipants) != 2 {
		t.Fatalf("expected UpdateGroupParticipants to be called with 2 JIDs, got %d", len(fake.UpdatedGroupParticipants))
	}
	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 message sent (the success reaction), got %d", len(fake.SentMessages))
	}
}

func TestBanUserPropagatesUpdateError(t *testing.T) {
	wantErr := errors.New("update failed")
	fake := &botfakes.FakeClient{
		UpdateGroupParticipantsFunc: func(ctx context.Context, jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantChange) ([]types.GroupParticipant, error) {
			return nil, wantErr
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.MentionedMembers = []*models.Member{{JID: "111@lid"}}

	BanUser(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "update failed") {
		t.Errorf("expected error reply to mention the failure, got %q", text)
	}
}

func TestWarnUserBelowThreshold(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	m.Chat.WarnBanThreshold = 4
	m.MentionedMembers = []*models.Member{{ChatID: m.Chat.ChatID, JID: "111@lid", Warns: 0}}

	WarnUser(m)

	if m.MentionedMembers[0].Warns != 1 {
		t.Errorf("expected Warns to be incremented to 1, got %d", m.MentionedMembers[0].Warns)
	}
	if len(fake.UpdatedGroupParticipants) != 0 {
		t.Error("expected no ban to happen below the threshold")
	}
	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "1 aviso") {
		t.Errorf("expected reply to mention the warn count, got %q", text)
	}
}

func TestWarnUserReachingThresholdBans(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	m.Chat.WarnBanThreshold = 1
	m.MentionedMembers = []*models.Member{{ChatID: m.Chat.ChatID, JID: "111@lid", Warns: 0}}

	WarnUser(m)

	if m.MentionedMembers[0].Warns != 0 {
		t.Errorf("expected Warns to reset to 0 after ban, got %d", m.MentionedMembers[0].Warns)
	}
	if len(fake.UpdatedGroupParticipants) != 1 {
		t.Errorf("expected the member to be banned once the threshold is reached, got %d ban calls", len(fake.UpdatedGroupParticipants))
	}
}

func TestRemoveUserWarnDecrementsWarns(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	m.MentionedMembers = []*models.Member{{ChatID: m.Chat.ChatID, JID: "111@lid", Warns: 2}}

	RemoveUserWarn(m)

	if m.MentionedMembers[0].Warns != 1 {
		t.Errorf("expected Warns to be decremented to 1, got %d", m.MentionedMembers[0].Warns)
	}
}

func TestRemoveUserWarnSkipsZeroWarns(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	m.MentionedMembers = []*models.Member{{ChatID: m.Chat.ChatID, JID: "111@lid", Warns: 0}}

	RemoveUserWarn(m)

	if m.MentionedMembers[0].Warns != 0 {
		t.Errorf("expected Warns to stay at 0, got %d", m.MentionedMembers[0].Warns)
	}
}

func TestMentionMembersRejectsEmptyMessage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.RawEvent.Message = &waE2E.Message{}

	MentionMembers(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + rejection reply), got %d", len(fake.SentMessages))
	}
	if len(fake.UpdatedGroupParticipants) != 0 {
		t.Error("expected no group calls for an empty message")
	}
}

func TestMentionMembersSuccess(t *testing.T) {
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{
				Participants: []types.GroupParticipant{
					{JID: types.NewJID("111", types.HiddenUserServer)},
					{JID: types.NewJID("222", types.HiddenUserServer)},
				},
			}, nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"hello", "everyone"}

	MentionMembers(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 messages sent (mention + reaction), got %d", len(fake.SentMessages))
	}
	ext := fake.SentMessages[0].Message.GetExtendedTextMessage()
	if ext.GetText() != "hello everyone" {
		t.Errorf("text = %q, want %q", ext.GetText(), "hello everyone")
	}
	if len(ext.GetContextInfo().GetMentionedJID()) != 2 {
		t.Errorf("expected 2 mentioned JIDs, got %d", len(ext.GetContextInfo().GetMentionedJID()))
	}
}

func TestMentionMembersGetGroupInfoError(t *testing.T) {
	wantErr := errors.New("group lookup failed")
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) { return nil, wantErr },
	}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"hello"}

	MentionMembers(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 reply (the error), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[0].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "group lookup failed") {
		t.Errorf("expected the reply to mention the error, got %q", text)
	}
}

func TestChangeUserAdminStatusPromote(t *testing.T) {
	fake := &botfakes.FakeClient{}
	var gotAction whatsmeow.ParticipantChange
	fake.UpdateGroupParticipantsFunc = func(ctx context.Context, jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantChange) ([]types.GroupParticipant, error) {
		gotAction = action
		return nil, nil
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.MentionedMembers = []*models.Member{{JID: "111@lid"}}

	if err := changeUserAdminStatus(m); err != nil {
		t.Fatalf("changeUserAdminStatus() error = %v", err)
	}
	if gotAction != whatsmeow.ParticipantChangePromote {
		t.Errorf("action = %v, want ParticipantChangePromote", gotAction)
	}
}

func TestChangeUserAdminStatusDemote(t *testing.T) {
	fake := &botfakes.FakeClient{}
	var gotAction whatsmeow.ParticipantChange
	fake.UpdateGroupParticipantsFunc = func(ctx context.Context, jid types.JID, participantChanges []types.JID, action whatsmeow.ParticipantChange) ([]types.GroupParticipant, error) {
		gotAction = action
		return nil, nil
	}
	m := newTestMessage(t, fake, newTestDB(t))
	m.MentionedMembers = []*models.Member{{JID: "111@lid"}}

	if err := changeUserAdminStatus(m, true); err != nil {
		t.Fatalf("changeUserAdminStatus() error = %v", err)
	}
	if gotAction != whatsmeow.ParticipantChangeDemote {
		t.Errorf("action = %v, want ParticipantChangeDemote", gotAction)
	}
}

func TestChangeUserAdminStatusInvalidJID(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.MentionedMembers = []*models.Member{{JID: "111:notanumber@lid"}}

	if err := changeUserAdminStatus(m); err == nil {
		t.Error("expected an error for an unparseable JID")
	}
}

func TestMessagesPerMemberDedupsAndSorts(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)

	for _, mem := range []*models.Member{
		{ChatID: m.Chat.ChatID, JID: "111@lid", Messages: 5},
		{ChatID: m.Chat.ChatID, JID: "222@lid", Messages: 10},
		{ChatID: m.Chat.ChatID, JID: "333@lid", Messages: 0},
	} {
		if _, err := db.Member.GetOrCreateMember(m.Chat.ChatID, mem.JID); err != nil {
			t.Fatalf("seed GetOrCreateMember() error = %v", err)
		}
		if err := db.Member.Update(mem); err != nil {
			t.Fatalf("seed Update() error = %v", err)
		}
	}

	MessagesPerMember(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	idx222 := strings.Index(text, "@222")
	idx111 := strings.Index(text, "@111")
	if idx222 == -1 || idx111 == -1 || idx222 > idx111 {
		t.Errorf("expected @222 (10 msgs) listed before @111 (5 msgs), got:\n%s", text)
	}
	if strings.Contains(text, "333") {
		t.Errorf("expected member with 0 messages to be omitted, got:\n%s", text)
	}
}

func TestPurgeMessagesResetsCount(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)

	if _, err := db.Member.GetOrCreateMember(m.Chat.ChatID, "111@lid"); err != nil {
		t.Fatalf("seed error = %v", err)
	}
	if err := db.Member.Update(&models.Member{ChatID: m.Chat.ChatID, JID: "111@lid", Messages: 42}); err != nil {
		t.Fatalf("seed Update() error = %v", err)
	}

	PurgeMessages(m)

	got, err := db.Member.GetOrCreateMember(m.Chat.ChatID, "111@lid")
	if err != nil {
		t.Fatalf("fetch error = %v", err)
	}
	if got.Messages != 0 {
		t.Errorf("expected Messages to be reset to 0, got %d", got.Messages)
	}
}

func TestGetMembersZeroMessages(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)

	for jid, msgs := range map[string]int{"111@lid": 0, "222@lid": 5} {
		if _, err := db.Member.GetOrCreateMember(m.Chat.ChatID, jid); err != nil {
			t.Fatalf("seed error = %v", err)
		}
		if err := db.Member.Update(&models.Member{ChatID: m.Chat.ChatID, JID: jid, Messages: msgs}); err != nil {
			t.Fatalf("seed Update() error = %v", err)
		}
	}

	GetMembersZeroMessages(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "@111") {
		t.Errorf("expected zero-message member to be listed, got:\n%s", text)
	}
	if strings.Contains(text, "@222") {
		t.Errorf("expected member with messages to be omitted, got:\n%s", text)
	}
}

func TestMuteMemberSetsSilenced(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	m.MentionedMembers = []*models.Member{{ChatID: m.Chat.ChatID, JID: "111@lid", Silenced: 0}}

	MuteMember(m)

	if m.MentionedMembers[0].Silenced != 1 {
		t.Errorf("expected Silenced to be set to 1, got %d", m.MentionedMembers[0].Silenced)
	}
	if len(fake.SentMessages) != 1 {
		t.Errorf("expected a success reaction, got %d messages", len(fake.SentMessages))
	}
}

func TestUnmuteMemberClearsSilenced(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	m.MentionedMembers = []*models.Member{{ChatID: m.Chat.ChatID, JID: "111@lid", Silenced: 1}}

	UnmuteMember(m)

	if m.MentionedMembers[0].Silenced != 0 {
		t.Errorf("expected Silenced to be cleared to 0, got %d", m.MentionedMembers[0].Silenced)
	}
}
