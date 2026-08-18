package admin

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func lastReplyText(fake *botfakes.FakeClient) string {
	if len(fake.SentMessages) == 0 {
		return ""
	}
	return fake.SentMessages[len(fake.SentMessages)-1].Message.GetExtendedTextMessage().GetText()
}

func TestSetupNoArgsShowsCurrentConfig(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	Setup(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "prefixo=") {
		t.Errorf("expected the current config to be echoed back, got %q", text)
	}
}

func TestSetupWithArgsAppliesConfig(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"prefixo=!"}

	Setup(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	if m.Chat.Prefix != "!" {
		t.Errorf("expected chat prefix to be updated to '!', got %q", m.Chat.Prefix)
	}
}

func TestSetupInvalidConfigReplies(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"chaveInvalida=sim"}

	Setup(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Opção não reconhecida") {
		t.Errorf("expected an unrecognized-option error, got %q", text)
	}
}

func TestBugNoCreatorConfigured(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.Bot.CreatorNumber = nil
	*m.Args = []string{"something", "broke"}

	Bug(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + missing-config reply) and no send attempt, got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "Nenhum numero configurado") {
		t.Errorf("expected a missing-config reply, got %q", text)
	}
}

func TestBugSendsToCreator(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	creator := "5511999999999"
	m.Bot.CreatorNumber = &creator
	*m.Args = []string{"something", "broke"}

	Bug(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (bug report + success reaction), got %d", len(fake.SentMessages))
	}
	reportText := fake.SentMessages[0].Message.GetConversation()
	if !strings.Contains(reportText, "something broke") {
		t.Errorf("expected the bug report to include the args, got %q", reportText)
	}
	wantJID := types.NewJID(creator, types.DefaultUserServer)
	if fake.SentMessages[0].Chat != wantJID {
		t.Errorf("expected the report to be sent to %v, got %v", wantJID, fake.SentMessages[0].Chat)
	}
}

func TestBugSendError(t *testing.T) {
	wantErr := errors.New("send failed")
	fake := &botfakes.FakeClient{
		SendMessageFunc: func(ctx context.Context, to types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
			return whatsmeow.SendResponse{}, wantErr
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	creator := "5511999999999"
	m.Bot.CreatorNumber = &creator
	*m.Args = []string{"something", "broke"}

	Bug(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (failed report + failed reaction), got %d", len(fake.SentMessages))
	}
}

func TestBroadcastMissingPasswordConfig(t *testing.T) {
	t.Setenv("BROADCAST_PASSWORD", "")
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"whatever", "message here"}

	Broadcast(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "Nenhuma senha") {
		t.Errorf("expected a missing-password reply, got %q", text)
	}
}

func TestBroadcastWrongPassword(t *testing.T) {
	t.Setenv("BROADCAST_PASSWORD", "correct-password")
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"wrong-password", "message here"}

	Broadcast(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Senha invalida") {
		t.Errorf("expected an invalid-password reply, got %q", text)
	}
}

func TestBroadcastGetJoinedGroupsError(t *testing.T) {
	t.Setenv("BROADCAST_PASSWORD", "correct-password")
	wantErr := errors.New("network error")
	fake := &botfakes.FakeClient{
		GetJoinedGroupsFunc: func(ctx context.Context) ([]*types.GroupInfo, error) { return nil, wantErr },
	}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"correct-password", "message", "here"}

	Broadcast(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "network error") {
		t.Errorf("expected the group-lookup error to be reported, got %q", text)
	}
}

func TestBroadcastSuccessReportsGroupCount(t *testing.T) {
	t.Setenv("BROADCAST_PASSWORD", "correct-password")
	fake := &botfakes.FakeClient{
		GetJoinedGroupsFunc: func(ctx context.Context) ([]*types.GroupInfo, error) {
			return []*types.GroupInfo{
				{JID: types.NewJID("g1", types.GroupServer)},
				{JID: types.NewJID("g2", types.GroupServer)},
			}, nil
		},
	}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"correct-password", "message", "here"}

	Broadcast(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + status reply) so far, got %d", len(fake.SentMessages))
	}
	text := lastReplyText(fake)
	if !strings.Contains(text, "2 grupos") {
		t.Errorf("expected the reply to report 2 groups, got %q", text)
	}
}
