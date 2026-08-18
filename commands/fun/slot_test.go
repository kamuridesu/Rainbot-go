package fun

import (
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
)

func TestSlotSendsAFormattedReply(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	Slot(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "SLOT") {
		t.Errorf("expected the slot machine art in the reply, got %q", text)
	}
	if !strings.Contains(text, "Parabéns") && !strings.Contains(text, "pena") {
		t.Errorf("expected either a win or lose message, got %q", text)
	}
}
