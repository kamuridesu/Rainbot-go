package offenses

import (
	"context"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func newTestMessage(text string, chat *models.Chat) *messages.Message {
	return &messages.Message{
		Chat: chat,
		Text: &text,
	}
}

func newTestMessageWithClient(fake *botfakes.FakeClient, text string, chat *models.Chat) *messages.Message {
	return &messages.Message{
		Ctx:  context.Background(),
		Bot:  &bot.Bot{Client: fake},
		Chat: chat,
		Text: &text,
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{Chat: types.NewJID("123", types.GroupServer)},
				ID:            "stanza-1",
			},
		},
	}
}

func TestOffendsBotDisabledByChatSetting(t *testing.T) {
	m := newTestMessage("bot inutil", &models.Chat{AllowOffensiveReplies: 0})
	if OffendsBot(m) {
		t.Error("expected false when AllowOffensiveReplies is disabled")
	}
}

func TestOffendsBotSkippedWhenProfanityFilterEnabled(t *testing.T) {
	m := newTestMessage("bot inutil", &models.Chat{AllowOffensiveReplies: 1, ProfanityFilterEnabled: 1})
	if OffendsBot(m) {
		t.Error("expected false when the profanity filter is enabled (it handles offenses itself)")
	}
}

func TestOffendsBotIgnoresTextWithoutBotMention(t *testing.T) {
	m := newTestMessage("isso e um lixo", &models.Chat{AllowOffensiveReplies: 1, ProfanityFilterEnabled: 0})
	if OffendsBot(m) {
		t.Error("expected false when text doesn't mention 'bot'")
	}
}

func TestOffendsBotIgnoresNonOffensiveBotMention(t *testing.T) {
	m := newTestMessage("bom dia bot, tudo bem?", &models.Chat{AllowOffensiveReplies: 1, ProfanityFilterEnabled: 0})
	if OffendsBot(m) {
		t.Error("expected false when 'bot' is mentioned without an offensive word")
	}
}

func TestOffendsBotRepliesToOffensiveMention(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessageWithClient(fake, "bot inutil", &models.Chat{AllowOffensiveReplies: 1, ProfanityFilterEnabled: 0})

	if !OffendsBot(m) {
		t.Fatal("expected true for an offensive bot mention")
	}
	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 reply (no reaction on this path), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[0].Message.GetExtendedTextMessage().GetText() == "" {
		t.Error("expected a non-empty comeback reply")
	}
}

func TestOffendsBotCaseInsensitive(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessageWithClient(fake, "BOT É LIXO", &models.Chat{AllowOffensiveReplies: 1, ProfanityFilterEnabled: 0})

	if !OffendsBot(m) {
		t.Error("expected true regardless of case")
	}
}
