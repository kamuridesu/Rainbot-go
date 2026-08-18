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
	err := ParseSetupText([]string{"noEqualsSign"}, chat, nil)
	if err == nil {
		t.Fatal("expected error for line without '='")
	}
}

func TestParseSetupTextRejectsUnknownKey(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"chaveInvalida=sim"}, chat, nil)
	if err == nil {
		t.Fatal("expected error for unrecognized key")
	}
}

func TestParseSetupTextRejectsInvalidNumber(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"limiteDeAvisos=abc"}, chat, nil)
	if err == nil {
		t.Fatal("expected error for non-numeric limiteDeAvisos")
	}
}

func TestParseSetupTextRejectsInvalidBoolValue(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"apenasAdmin=talvez"}, chat, nil)
	if err == nil {
		t.Fatal("expected error for non sim/nao value")
	}
}

func TestParseSetupTextRejectsMultiCharPrefix(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"prefixo=ab"}, chat, nil)
	if err == nil {
		t.Fatal("expected error for a prefix longer than one character")
	}
}

func TestParseSetupTextSetsFieldsBeforeFailing(t *testing.T) {
	chat := &models.Chat{}
	err := ParseSetupText([]string{"prefixo=!", "chaveInvalida=sim"}, chat, nil)
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
		AllowGames:             1,
		AllowAdults:            0,
		AllowFun:               1,
		CountMessages:          0,
		ProfanityFilterEnabled: 1,
		CustomProfanityWords:   "foo,bar",
		WelcomeMessage:         "welcome!",
		AllowQuote:             1,
		QuoteNMessages:         300,
		AllowOffensiveReplies:  0,
	}

	got := GetHumanReadableSetup(chat)

	wantContains := []string{
		"prefixo=!",
		"ativarBot=sim",
		"apenasAdmin=não",
		"limiteDeAvisos=3",
		"ativarGames=sim",
		"ativarAdultos=não",
		"ativarDiversao=sim",
		"contarMensagens=não",
		"filtroDeProfanidade=sim",
		"palavrasProibidas=foo,bar",
		`boasVindas="welcome!"`,
		"ativarQuote=sim",
		"quoteRate=300",
		"responderOfensa=não",
	}

	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("GetHumanReadableSetup() missing %q, got:\n%s", want, got)
		}
	}
}
