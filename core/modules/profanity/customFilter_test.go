package profanity

import (
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
)

func TestCheckCustomWord(t *testing.T) {
	tests := []struct {
		name    string
		chat    *models.Chat
		text    string
		wantErr bool
	}{
		{
			name:    "no custom words configured",
			chat:    &models.Chat{CustomProfanityWords: ""},
			text:    "anything goes here",
			wantErr: false,
		},
		{
			name:    "text contains a configured word",
			chat:    &models.Chat{CustomProfanityWords: "foo,bar"},
			text:    "say foo now",
			wantErr: true,
		},
		{
			name:    "text does not contain any configured word",
			chat:    &models.Chat{CustomProfanityWords: "foo,bar"},
			text:    "hello world",
			wantErr: false,
		},
		{
			name:    "match is case insensitive",
			chat:    &models.Chat{CustomProfanityWords: "foo"},
			text:    "say FOO now",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckCustomWord(tt.chat, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckCustomWord(%q) error = %v, wantErr %v", tt.text, err, tt.wantErr)
			}
		})
	}
}
