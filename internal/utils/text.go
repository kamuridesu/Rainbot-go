package utils

import (
	"regexp"
)

type MentionMessage struct {
	Text    string
	Mention []string
}

func GenerateMentionFromText(text string) *MentionMessage {

	re := regexp.MustCompile(`(\d+)@lid`)
	lidsSlice := re.FindAllString(text, -1)
	replaced := re.ReplaceAllString(text, "@$1")

	return &MentionMessage{
		Text:    replaced,
		Mention: lidsSlice,
	}

}
