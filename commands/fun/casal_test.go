package fun

import (
	"context"
	"errors"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
)

func TestCasalGetGroupInfoError(t *testing.T) {
	wantErr := errors.New("group lookup failed")
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) { return nil, wantErr },
	}
	m := newTestMessage(t, fake, newTestDB(t))

	Casal(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + error reply), got %d", len(fake.SentMessages))
	}
}

func TestCasalPicksTwoMembersExcludingBot(t *testing.T) {
	botJID := types.NewJID("999", types.HiddenUserServer)
	fake := &botfakes.FakeClient{
		GetGroupInfoFunc: func(ctx context.Context, jid types.JID) (*types.GroupInfo, error) {
			return &types.GroupInfo{
				Participants: []types.GroupParticipant{
					{LID: botJID},
					{LID: types.NewJID("111", types.HiddenUserServer)},
					{LID: types.NewJID("222", types.HiddenUserServer)},
				},
			}, nil
		},
		StoreFunc: func() *store.Device { return &store.Device{LID: botJID} },
	}
	m := newTestMessage(t, fake, newTestDB(t))

	Casal(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (couple announcement + reaction), got %d", len(fake.SentMessages))
	}
	mentions := fake.SentMessages[0].Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
	if len(mentions) != 2 {
		t.Fatalf("expected 2 mentioned JIDs, got %d", len(mentions))
	}
	for _, mention := range mentions {
		if mention == botJID.String() {
			t.Error("expected the bot itself to never be picked as part of the couple")
		}
	}
}
