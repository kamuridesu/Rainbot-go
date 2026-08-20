package fun

import (
	"bytes"
	"image/jpeg"
	"strings"
	"testing"
)

func TestGenerateRucoyUpskillCard(t *testing.T) {
	card, err := generateRucoyUpskillCard(RucoyUpskillCardData{
		FromSkill:     400,
		ToSkill:       450,
		EstimatedTime: "26 horas e 30 minutos",
		DailyHours:    8,
		Options: RucoyUpskillOptions{
			DailyHours:   8,
			Vocation:     "Kina",
			ManaPerSkill: 50,
		},
		ManaEstimate: RucoyUpskillManaEstimate{
			TotalMana:  6625000,
			MinPotions: 7362,
			MaxPotions: 11042,
			MinCost:    4810000,
			MaxCost:    7280000,
		},
	})
	if err != nil {
		t.Fatalf("generateRucoyUpskillCard() error = %v", err)
	}
	if len(card) == 0 {
		t.Fatal("expected generated card bytes")
	}
	if _, err := jpeg.Decode(bytes.NewReader(card)); err != nil {
		t.Fatalf("generated card is not a valid jpeg: %v", err)
	}
}

func TestCompactRucoyDuration(t *testing.T) {
	if got := compactRucoyDuration("26 horas e 30 minutos"); got != "26H 30MIN" {
		t.Errorf("compactRucoyDuration() = %q, want %q", got, "26H 30MIN")
	}

	got := compactRucoyDailyDuration("26 horas e 30 minutos", 8)
	if !strings.Contains(got, "8H/DIA: 3D 2H 30MIN") {
		t.Errorf("compactRucoyDailyDuration() = %q", got)
	}
}

func TestFormatRucoyCardNumber(t *testing.T) {
	tests := []struct {
		value int64
		want  string
	}{
		{999, "999"},
		{596362, "596K"},
		{536725000, "536.7KK"},
		{581000000, "581KK"},
	}

	for _, tt := range tests {
		if got := formatRucoyCardNumber(tt.value); got != tt.want {
			t.Errorf("formatRucoyCardNumber(%d) = %q, want %q", tt.value, got, tt.want)
		}
	}
}
