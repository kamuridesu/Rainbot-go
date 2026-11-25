package llm

import (
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func newMatchMessage(text string, group bool, quoted *messages.Message) *messages.Message {
	name := "Rainbot"
	return &messages.Message{
		Bot:           &bot.Bot{Name: &name},
		Text:          &text,
		Chat:          &models.Chat{AllowLLM: 1},
		QuotedMessage: quoted,
		RawEvent:      &events.Message{Info: types.MessageInfo{MessageSource: types.MessageSource{IsGroup: group}}},
	}
}

func TestMatches(t *testing.T) {
	replyToBot := &messages.Message{IsFromMe: true}

	tests := []struct {
		name   string
		text   string
		group  bool
		quoted *messages.Message
		want   bool
	}{
		{"private text", "oi", false, nil, true},
		{"private empty (reaction/protocol)", "", false, nil, false},
		{"private whitespace", "  \n", false, nil, false},
		{"private empty reply to bot", "", false, replyToBot, false},
		{"group without mention", "oi gente", true, nil, false},
		{"group mention alias", "oi hifumi", true, nil, true},
		{"group mention name", "rainbot, tudo bem?", true, nil, true},
		{"group reply to bot", "e ai", true, replyToBot, true},
		{"group empty", "", true, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Matches(newMatchMessage(tt.text, tt.group, tt.quoted)); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("llm disabled", func(t *testing.T) {
		m := newMatchMessage("oi", false, nil)
		m.Chat.AllowLLM = 0
		if Matches(m) {
			t.Error("Matches() = true with AllowLLM = 0")
		}
	})
}
