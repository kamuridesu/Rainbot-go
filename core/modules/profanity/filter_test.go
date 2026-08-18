package profanity

import "testing"

func TestHasObsceneWord(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"clean sentence", "bom dia pessoal, tudo bem?", false},
		{"contains a banned word", "seu babaca chato", true},
		{"is case insensitive", "SEU BABACA", true},
		{"substring match does not trigger", "abacate esta gostoso", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HasObsceneWord(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasObsceneWord(%q) error = %v, wantErr %v", tt.text, err, tt.wantErr)
			}
		})
	}
}
