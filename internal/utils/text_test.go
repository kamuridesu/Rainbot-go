package utils

import (
	"reflect"
	"testing"
)

func TestParseLidToMention(t *testing.T) {
	tests := []struct {
		name string
		lid  string
		want string
	}{
		{"strips server suffix", "5511999999999@lid", "@5511999999999"},
		{"no suffix present", "5511999999999", "@5511999999999"},
		{"whatsapp.net suffix", "5511999999999@s.whatsapp.net", "@5511999999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseLidToMention(tt.lid); got != tt.want {
				t.Errorf("ParseLidToMention(%q) = %q, want %q", tt.lid, got, tt.want)
			}
		})
	}
}

func TestGenerateMentionFromText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantText string
		wantMent []string
	}{
		{
			name:     "no mentions",
			text:     "hello world",
			wantText: "hello world",
			wantMent: nil,
		},
		{
			name:     "at-prefixed mention",
			text:     "hello @5511999999999 how are you",
			wantText: "hello @5511999999999 how are you",
			wantMent: []string{"5511999999999@lid"},
		},
		{
			name:     "lid-suffixed mention gets rewritten to @ form",
			text:     "hello 5511999999999@lid how are you",
			wantText: "hello @5511999999999 how are you",
			wantMent: []string{"5511999999999@lid"},
		},
		{
			name:     "multiple mentions of both forms",
			text:     "@111 and 222@lid",
			wantText: "@111 and @222",
			wantMent: []string{"111@lid", "222@lid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateMentionFromText(tt.text)
			if got.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", got.Text, tt.wantText)
			}
			if !reflect.DeepEqual(got.Mention, tt.wantMent) {
				t.Errorf("Mention = %v, want %v", got.Mention, tt.wantMent)
			}
		})
	}
}

func TestParseArgsFromMessage(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"empty string", "", nil},
		{"single word", "hello", []string{"hello"}},
		{"multiple words", "hello world foo", []string{"hello", "world", "foo"}},
		{"collapses repeated spaces", "hello   world", []string{"hello", "world"}},
		{"splits on newlines", "hello\nworld", []string{"hello", "world"}},
		{"quoted segment kept as one arg", `say "hello world" now`, []string{"say", "hello world", "now"}},
		{"unterminated quote consumes rest as one token", `say "hello world`, []string{"say", "hello world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseArgsFromMessage(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArgsFromMessage(%q) = %#v, want %#v", tt.text, got, tt.want)
			}
		})
	}
}
