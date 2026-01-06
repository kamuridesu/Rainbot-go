package utils

import (
	"regexp"
	"strings"
)

type MentionMessage struct {
	Text    string
	Mention []string
}

func ParseLidToMention(lid string) string {
	s := strings.Split(lid, "@")[0]
	return "@" + s
}

func GenerateMentionFromText(text string) *MentionMessage {

	re := regexp.MustCompile(`@(\d+)`)
	lidsSlice := re.FindAllString(text, -1)
	replaced := re.ReplaceAllString(text, "@$1")

	for i, s := range lidsSlice {
		lidsSlice[i] = strings.Replace(s, "@", "", 1) + "@lid"
	}

	if len(lidsSlice) == 0 {
		lidsSlice = nil
	}

	return &MentionMessage{
		Text:    replaced,
		Mention: lidsSlice,
	}

}

func ParseArgsFromMessage(text string) []string {
	var parts []string
	isInsideQuotes := false
	curr := ""
	for _, r := range text {
		if r == '"' {
			isInsideQuotes = !isInsideQuotes
		} else if (r == ' ' || r == '\n') && !isInsideQuotes {
			if curr != "" {
				parts = append(parts, curr)
				curr = ""
			}
		} else {
			curr += string(r)
		}
	}
	if curr != "" {
		parts = append(parts, curr)
	}
	return parts
}
