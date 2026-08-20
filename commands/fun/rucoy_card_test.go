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
			TotalMana:  4770000,
			MinPotions: 5300,
			MaxPotions: 7950,
			MinCost:    3510000,
			MaxCost:    5200000,
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

func TestGenerateRucoyUpskillCardForPally(t *testing.T) {
	card, err := generateRucoyUpskillCard(RucoyUpskillCardData{
		FromSkill:     400,
		ToSkill:       450,
		EstimatedTime: "26 horas e 30 minutos",
		DailyHours:    8,
		Options: RucoyUpskillOptions{
			DailyHours:   8,
			Vocation:     "Pally",
			ManaPerSkill: 50,
		},
		ManaEstimate: RucoyUpskillManaEstimate{
			TotalMana:   4770000,
			MinPotions:  5300,
			MaxPotions:  7950,
			TotalArrows: 381600,
			ArrowCost:   764000,
			MinCost:     4274000,
			MaxCost:     5964000,
		},
	})
	if err != nil {
		t.Fatalf("generateRucoyUpskillCard() pally error = %v", err)
	}
	if len(card) == 0 {
		t.Fatal("expected generated pally card bytes")
	}
	if _, err := jpeg.Decode(bytes.NewReader(card)); err != nil {
		t.Fatalf("generated pally card is not a valid jpeg: %v", err)
	}
}

func TestGenerateRucoyUpskillCardForMage(t *testing.T) {
	card, err := generateRucoyUpskillCard(RucoyUpskillCardData{
		FromSkill:     400,
		ToSkill:       450,
		EstimatedTime: "26 horas e 30 minutos",
		DailyHours:    8,
		Options: RucoyUpskillOptions{
			DailyHours:   8,
			Vocation:     "Mage",
			ManaPerSkill: 40,
		},
		ManaEstimate: RucoyUpskillManaEstimate{
			TotalMana:  3816000,
			MinPotions: 4240,
			MaxPotions: 6360,
			MinCost:    2860000,
			MaxCost:    4160000,
		},
	})
	if err != nil {
		t.Fatalf("generateRucoyUpskillCard() mage error = %v", err)
	}
	if len(card) == 0 {
		t.Fatal("expected generated mage card bytes")
	}
	if _, err := jpeg.Decode(bytes.NewReader(card)); err != nil {
		t.Fatalf("generated mage card is not a valid jpeg: %v", err)
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

func TestFitPixelTextScaleUsesVisualWidth(t *testing.T) {
	text := "UPSKILL RUCOY"
	maxWidth := pixelTextWidth(text, 5)

	if pixelTextVisualWidth(text, 5) <= maxWidth {
		t.Fatal("expected visual width to include the shadow offset")
	}
	scale := fitPixelTextScale(text, maxWidth, 5)
	if scale >= 5 {
		t.Fatalf("expected scale to shrink when the shadow would exceed max width, got %d", scale)
	}
	if got := pixelTextVisualWidth(text, scale); got > maxWidth {
		t.Fatalf("fitted visual width = %d, want <= %d", got, maxWidth)
	}
}
