package llm

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/chatter"
)

func hasBotMention(text string, names ...string) bool {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	for _, w := range words {
		if w == "bot" {
			return true
		}
		for _, n := range names {
			if w == strings.ToLower(n) {
				return true
			}
		}
	}
	return false
}

func Answer(m *messages.Message) error {
	if m.Chat.AllowLLM == 0 {
		return nil
	}

	sendMessage := func(message string) error {
		c := chatter.New(m.Bot.DB.AIModel)

		answer, err := c.GenerateMessage(m.Ctx, *m.Bot.Name, message)
		if err != nil {
			return err
		}
		m.Reply(answer)
		return nil
	}

	if hasBotMention(*m.Text, "hifumi") {
		return sendMessage(*m.Text + "\nauthor: " + m.Author.JID)
	}

	if m.QuotedMessage == nil {
		return errors.New("quoted mesage is empty")
	}

	if !m.QuotedMessage.IsFromMe {
		return errors.New("quoted message is not from bot")
	}

	var text strings.Builder
	fmt.Fprintf(&text, "Quoted Message: %s; Author: %s\n\nCurrentMessage: %s; Author: %s", *m.QuotedMessage.Text, *m.Bot.Name, *m.Text, m.Author.JID)

	return sendMessage(text.String())
}
