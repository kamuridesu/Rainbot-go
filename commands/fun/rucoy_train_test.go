package fun

import (
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
)

func TestRucoyTrainSuccess(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"5", "351", "391", "-50"}

	RucoyTrain(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Setup") {
		t.Errorf("expected a training report, got %q", text)
	}
}

func TestRucoyTrainInvalidArgs(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"notanumber", "351", "391", "-50"}

	RucoyTrain(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "numero") {
		t.Errorf("expected a validation error reply, got %q", text)
	}
}

func TestRucoyTrainOutOfRangeInputRejected(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"5", "0", "391", "-50"}

	RucoyTrain(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "level") {
		t.Errorf("expected a level-range validation reply, got %q", text)
	}
}

func TestRucoyTrainWithCustomEfficiency(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"5", "351", "391", "-50", "90"}

	RucoyTrain(m)

	text := lastReplyText(fake)
	if !strings.Contains(text, "Eficiencia alvo: 90%") {
		t.Errorf("expected the custom efficiency to be reflected in the reply, got %q", text)
	}
}
