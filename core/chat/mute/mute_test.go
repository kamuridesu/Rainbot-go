package mute

import (
	"context"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func newTestMessage(fake *botfakes.FakeClient, silenced int) *messages.Message {
	return &messages.Message{
		Ctx: context.Background(),
		Bot: &bot.Bot{Client: fake},
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Chat:   types.NewJID("123", types.GroupServer),
					Sender: types.NewJID("111", types.HiddenUserServer),
				},
				ID: "stanza-1",
			},
		},
		Author: &models.Member{Silenced: silenced},
	}
}

func TestDeleteIfMutedDeletesMessageFromSilencedMember(t *testing.T) {
	var buildRevokeCalled bool
	fake := &botfakes.FakeClient{
		BuildRevokeFunc: func(chat, sender types.JID, id types.MessageID) *waE2E.Message {
			buildRevokeCalled = true
			return &waE2E.Message{}
		},
	}
	m := newTestMessage(fake, 1)

	if !DeleteIfMuted(m) {
		t.Error("expected DeleteIfMuted() to return true for a silenced member")
	}
	if !buildRevokeCalled {
		t.Error("expected BuildRevoke() to be called")
	}
	if len(fake.SentMessages) != 1 {
		t.Errorf("expected 1 sent message (the revoke), got %d", len(fake.SentMessages))
	}
}

func TestDeleteIfMutedIgnoresUnsilencedMember(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake, 0)

	if DeleteIfMuted(m) {
		t.Error("expected DeleteIfMuted() to return false for a non-silenced member")
	}
	if len(fake.SentMessages) != 0 {
		t.Error("expected no messages to be sent")
	}
}
