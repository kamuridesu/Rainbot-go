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
	re := regexp.MustCompile(`@(\d+)|(\d+)@lid`)

	matches := re.FindAllStringSubmatch(text, -1)

	var lidsSlice []string

	for _, match := range matches {
		var id string

		if match[1] != "" {
			id = match[1]
		}
		if match[2] != "" {
			id = match[2]
		}

		if id != "" {
			lidsSlice = append(lidsSlice, id+"@lid")
		}
	}

	replaced := text
	reReplace := regexp.MustCompile(`\b(\d+)@lid\b`)
	replaced = reReplace.ReplaceAllString(replaced, "@$1")

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
