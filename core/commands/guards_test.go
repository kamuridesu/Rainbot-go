package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func newGroupMessage(fake *botfakes.FakeClient, isGroup bool, authorJID string) *messages.Message {
	chatJID := types.NewJID("123456", types.GroupServer)
	return &messages.Message{
		Ctx: context.Background(),
		Bot: &bot.Bot{Client: fake},
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{Chat: chatJID, IsGroup: isGroup},
			},
		},
		Author: &models.Member{JID: authorJID},
	}
}

func TestIsGroup(t *testing.T) {
	if err := IsGroup(newGroupMessage(&botfakes.FakeClient{}, true, "")); err != nil {
		t.Errorf("IsGroup() in a group chat, error = %v", err)
	}
	if err := IsGroup(newGroupMessage(&botfakes.FakeClient{}, false, "")); err == nil {
		t.Error("expected an error for a non-group chat")
	}
}

func TestIsAdminRejectsNonGroupChats(t *testing.T) {
	m := newGroupMessage(&botfakes.FakeClient{}, false, "111@lid")
	if err := IsAdmin(m); err == nil {
		t.Error("expected an error for a non-group chat")
	}
}

func TestIsAdminPropagatesGetGroupInfoError(t *testing.T) {
	wantErr := errors.New("network error")
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) { return nil, wantErr },
	}
	m := newGroupMessage(fake, true, "111@lid")

	err := IsAdmin(m)
	if err == nil {
		t.Fatal("expected an error when GetGroupInfo fails")
	}
}

func TestIsAdminAllowsGroupAdmin(t *testing.T) {
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{
				Participants: []types.GroupParticipant{
					{LID: types.NewJID("111", types.HiddenUserServer), IsAdmin: true},
				},
			}, nil
		},
	}
	m := newGroupMessage(fake, true, "111@lid")

	if err := IsAdmin(m); err != nil {
		t.Errorf("IsAdmin() for a group admin, error = %v", err)
	}
}

func TestIsAdminRejectsNonAdminMember(t *testing.T) {
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{
				Participants: []types.GroupParticipant{
					{LID: types.NewJID("111", types.HiddenUserServer), IsAdmin: false},
				},
			}, nil
		},
	}
	m := newGroupMessage(fake, true, "111@lid")

	if err := IsAdmin(m); err == nil {
		t.Error("expected an error for a non-admin member")
	}
}

func TestIsAdminRejectsMemberNotInGroup(t *testing.T) {
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{Participants: nil}, nil
		},
	}
	m := newGroupMessage(fake, true, "111@lid")

	if err := IsAdmin(m); err == nil {
		t.Error("expected an error when the author isn't in the participant list")
	}
}

func TestIsBotAdminPropagatesGetGroupInfoError(t *testing.T) {
	wantErr := errors.New("network error")
	fake := &botfakes.FakeClient{
		StoreFunc:        func() *store.Device { return &store.Device{LID: types.NewJID("999", types.HiddenUserServer)} },
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) { return nil, wantErr },
	}
	m := newGroupMessage(fake, true, "")

	if err := IsBotAdmin(m); err == nil {
		t.Fatal("expected an error when GetGroupInfo fails")
	}
}

func TestIsBotAdminTrueWhenBotIsAdmin(t *testing.T) {
	fake := &botfakes.FakeClient{
		StoreFunc: func() *store.Device { return &store.Device{LID: types.NewJID("999", types.HiddenUserServer)} },
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{
				Participants: []types.GroupParticipant{
					{LID: types.NewJID("999", types.HiddenUserServer), IsAdmin: true},
				},
			}, nil
		},
	}
	m := newGroupMessage(fake, true, "")

	if err := IsBotAdmin(m); err != nil {
		t.Errorf("IsBotAdmin() when the bot is an admin, error = %v", err)
	}
}

func TestIsBotAdminFalseWhenBotIsNotAdmin(t *testing.T) {
	fake := &botfakes.FakeClient{
		StoreFunc: func() *store.Device { return &store.Device{LID: types.NewJID("999", types.HiddenUserServer)} },
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{
				Participants: []types.GroupParticipant{
					{LID: types.NewJID("999", types.HiddenUserServer), IsAdmin: false},
				},
			}, nil
		},
	}
	m := newGroupMessage(fake, true, "")

	if err := IsBotAdmin(m); err == nil {
		t.Error("expected an error when the bot is not an admin")
	}
}

func TestHasArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		numArgs int
		atMax   bool
		wantErr bool
	}{
		{"zero required always passes", 0, false, nil, false},
		{"enough args", 2, false, []string{"a", "b"}, false},
		{"more than required, not atMax", 2, false, []string{"a", "b", "c"}, false},
		{"too few args", 2, false, []string{"a"}, true},
		{"exact match with atMax", 2, true, []string{"a", "b"}, false},
		{"too many args with atMax", 2, true, []string{"a", "b", "c"}, true},
		{"too few args with atMax", 2, true, []string{"a"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &messages.Message{Args: &tt.args}
			guard := HasArgs(tt.numArgs, tt.atMax)
			err := guard(m)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasArgs(%d, %v)(args=%v) error = %v, wantErr %v", tt.numArgs, tt.atMax, tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestHasArgsWithoutAtMaxVariadic(t *testing.T) {
	args := []string{"a"}
	m := &messages.Message{Args: &args}

	if err := HasArgs(2)(m); err == nil {
		t.Error("expected error when fewer args than required and atMax omitted")
	}
}

func TestHasMentionedMembers(t *testing.T) {
	t.Run("no mentions", func(t *testing.T) {
		m := &messages.Message{MentionedMembers: nil}
		if err := HasMentionedMembers(m); err == nil {
			t.Error("expected error when no members are mentioned")
		}
	})

	t.Run("has mentions", func(t *testing.T) {
		m := &messages.Message{MentionedMembers: []*models.Member{{JID: "111@lid"}}}
		if err := HasMentionedMembers(m); err != nil {
			t.Errorf("expected no error when a member is mentioned, got %v", err)
		}
	})
}

func TestHasQuotedMessage(t *testing.T) {
	t.Run("no quoted message", func(t *testing.T) {
		m := &messages.Message{QuotedMessage: nil}
		if err := HasQuotedMessage(m); err == nil {
			t.Error("expected error when there's no quoted message")
		}
	})

	t.Run("has quoted message", func(t *testing.T) {
		m := &messages.Message{QuotedMessage: &messages.Message{}}
		if err := HasQuotedMessage(m); err != nil {
			t.Errorf("expected no error when a message is quoted, got %v", err)
		}
	})
}
