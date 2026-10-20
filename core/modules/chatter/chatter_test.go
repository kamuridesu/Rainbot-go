package chatter

import (
	"strings"
	"testing"
)

func TestGetSystemPromptUsesDefaultStyleWhenEmpty(t *testing.T) {
	got := GetSystemPrompt("bot", "/", "")
	if !strings.HasSuffix(got, DefaultStyle) {
		t.Errorf("expected prompt to end with the default style, got:\n%s", got)
	}
}

func TestGetSystemPromptUsesCustomStyle(t *testing.T) {
	got := GetSystemPrompt("bot", "/", "seja formal")
	if !strings.HasSuffix(got, "seja formal") {
		t.Errorf("expected prompt to end with the custom style, got:\n%s", got)
	}
	if strings.Contains(got, DefaultStyle) {
		t.Error("custom style prompt should not include the default style")
	}
}
