package fun

import (
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
)

func TestChanceDeVirgindadeAlwaysZero(t *testing.T) {
	for _, word := range []string{"virgindade", "virgem"} {
		t.Run(word, func(t *testing.T) {
			fake := &botfakes.FakeClient{}
			m := newTestMessage(t, fake, newTestDB(t))
			*m.Args = []string{"de", "eu", "ser", word}

			ChanceDe(m)

			if len(fake.SentMessages) != 1 {
				t.Fatalf("expected 1 reply (no reaction on this path), got %d", len(fake.SentMessages))
			}
			text := fake.SentMessages[0].Message.GetExtendedTextMessage().GetText()
			if text != "Nenhuma" {
				t.Errorf("text = %q, want %q", text, "Nenhuma")
			}
		})
	}
}

func TestChanceDeGeneratesAPercentage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"eu", "ficar", "rico"}

	ChanceDe(m)

	text := fake.SentMessages[0].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "eu ficar rico") || !strings.Contains(text, "%") {
		t.Errorf("expected a percentage reply mentioning the args, got %q", text)
	}
}

func TestPercent(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"pobre"}

	Percent(m)

	text := fake.SentMessages[0].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(text, "pobre") || !strings.Contains(text, "%") {
		t.Errorf("expected a percentage reply mentioning 'pobre', got %q", text)
	}
}

func TestGado(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	Gado(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if !strings.HasPrefix(text, "Você é ") {
		t.Errorf("text = %q, want prefix %q", text, "Você é ")
	}
}

func TestGay(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	Gay(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + reply), got %d", len(fake.SentMessages))
	}
	text := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText()
	if text == "" {
		t.Error("expected a non-empty reply")
	}
}
