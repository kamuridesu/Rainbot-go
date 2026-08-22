package rucoy

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
	if !strings.Contains(text, "Arma 5 | Lvl 351 | Skill 391 | Add -50") {
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
	if !strings.Contains(text, "Alvo 90%+") {
		t.Errorf("expected the custom efficiency to be reflected in the reply, got %q", text)
	}
}

func TestRucoyTrainCleanOutputForSuggestedSetup(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"5", "354", "402", "-50"}

	RucoyTrain(m)

	want := `Arma 5 | Lvl 354 | Skill 402 | Add -50
Efetiva 352 | Alvo 90%+

AFK Train
Sem opção viável.
Melhor: Drow Fighter Lv.135 | 05:02
Muito curto para AFK Train.
Próximo: Lizard Archer Lv.160 | falta +54

Sugestão: arma 9
Dragon Hatchling Lv.240 | 09:07 | 99.8%

Power Train
Sem opção viável.
Melhor: Lizard Captain Lv.180 | 02:32
Muito curto para Power Train.
Próximo: Minotaur Lv.225 | falta +166

Sugestão: arma 11
Minotaur Lv.275 | 69:46 | 96.8%`

	if got := lastReplyText(fake); got != want {
		t.Errorf("unexpected clean train output:\n got:\n%q\nwant:\n%q", got, want)
	}
}
