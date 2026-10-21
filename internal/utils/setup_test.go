package utils

import (
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestValidateBool(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"sim", false},
		{"nao", false},
		{"não", false},
		{"yes", true},
		{"", true},
	}

	for _, tt := range tests {
		if err := validateBool(tt.value, 0); (err != nil) != tt.wantErr {
			t.Errorf("validateBool(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
		}
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("boolToInt(true) should be 1")
	}
	if boolToInt(false) != 0 {
		t.Error("boolToInt(false) should be 0")
	}
}

func TestParseSetupTextRejectsMalformedLine(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"noEqualsSign"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error for line without '='")
	}
}

func TestParseSetupTextRejectsUnknownKey(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"chaveInvalida=sim"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error for unrecognized key")
	}
}

func TestParseSetupTextRejectsInvalidNumber(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"limiteDeAvisos=abc"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error for non-numeric limiteDeAvisos")
	}
}

func TestParseSetupTextRejectsInvalidBoolValue(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"apenasAdmin=talvez"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error for non sim/nao value")
	}
}

func TestParseSetupTextRejectsMultiCharPrefix(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"prefixo=ab"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error for a prefix longer than one character")
	}
}

func TestParseSetupTextSetsFieldsBeforeFailing(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"prefixo=!", "chaveInvalida=sim"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error from the unknown key")
	}
	if chat.Prefix != "!" {
		t.Errorf("expected prefixo to be applied before the error, got %q", chat.Prefix)
	}
}

func TestGetHumanReadableSetup(t *testing.T) {
	chat := &models.Chat{
		Prefix:                 "!",
		IsBotEnabled:           1,
		AdminOnly:              0,
		WarnBanThreshold:       3,
		DisabledCategories:     "rucoy",
		CountMessages:          0,
		ProfanityFilterEnabled: 1,
		CustomProfanityWords:   "foo,bar",
		WelcomeMessage:         "welcome!",
		AllowQuote:             1,
		QuoteNMessages:         300,
		AllowOffensiveReplies:  0,
		AllowLLM:               0,
		BotStyle:               "seja formal",
	}

	got := GetHumanReadableSetup(chat, []string{"diversão", "rucoy"})

	wantContains := []string{
		"prefixo=!",
		"ativarBot=sim",
		"apenasAdmin=não",
		"limiteDeAvisos=3",
		"categoriasAtivas=diversão\n",
		"contarMensagens=não",
		"filtroDeProfanidade=sim",
		"palavrasProibidas=foo,bar",
		`boasVindas="welcome!"`,
		"ativarQuote=sim",
		"quoteRate=300",
		"responderOfensa=não",
		"ativarIA=não",
		`estiloBot="seja formal"`,
	}

	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("GetHumanReadableSetup() missing %q, got:\n%s", want, got)
		}
	}
}

func TestParseSetupTextBotStyle(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"estiloBot= Seja Formal ", "chaveInvalida=sim"}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error from the unknown key")
	}
	if chat.BotStyle != "Seja Formal" {
		t.Errorf("BotStyle = %q, want %q", chat.BotStyle, "Seja Formal")
	}
}

func TestParseSetupTextRejectsLongBotStyle(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"estiloBot=" + strings.Repeat("a", maxBotStyleLength+1)}, chat, nil, nil)
	if err == nil {
		t.Fatal("expected error for bot style over the limit")
	}
}

func TestParseSetupTextActiveCategories(t *testing.T) {
	chat := &models.Chat{}
	categories := []string{"diversão", "rucoy", "test"}
	err := ParseSetupText([]string{"categoriasAtivas=Diversão, test", "chaveInvalida=sim"}, chat, nil, categories)
	if err == nil {
		t.Fatal("expected error from the unknown key")
	}
	if chat.DisabledCategories != "rucoy" {
		t.Errorf("DisabledCategories = %q, want %q", chat.DisabledCategories, "rucoy")
	}
}

func TestParseSetupTextEmptyActiveCategoriesDisablesAll(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"categoriasAtivas=", "chaveInvalida=sim"}, chat, nil, []string{"diversão", "rucoy"})
	if err == nil {
		t.Fatal("expected error from the unknown key")
	}
	if chat.DisabledCategories != "diversão,rucoy" {
		t.Errorf("DisabledCategories = %q, want %q", chat.DisabledCategories, "diversão,rucoy")
	}
}

func TestParseSetupTextRejectsUnknownCategory(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"categoriasAtivas=diversão,admin"}, chat, nil, []string{"diversão", "rucoy"})
	if err == nil {
		t.Fatal("expected error for a category that can't be toggled")
	}
}
