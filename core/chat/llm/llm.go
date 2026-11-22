package llm

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/chatter"
)

var botAliases = []string{"bot", "hifumi"}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func isWordSeparator(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}

func hasBotMention(text string, names ...string) bool {
	for _, word := range strings.FieldsFunc(text, isWordSeparator) {
		for _, name := range names {
			if name != "" && strings.EqualFold(word, name) {
				return true
			}
		}
	}
	return false
}

func mentionsBot(m *messages.Message) bool {
	names := append([]string{deref(m.Bot.Name)}, botAliases...)
	return hasBotMention(deref(m.Text), names...)
}

func isReplyToBot(m *messages.Message) bool {
	return m.QuotedMessage != nil && m.QuotedMessage.IsFromMe
}

func Matches(m *messages.Message) bool {
	if m.Chat.AllowLLM == 0 {
		return false
	}
	return mentionsBot(m) || isReplyToBot(m)
}

func writeMessage(b *strings.Builder, label string, author any, content string) {
	fmt.Fprintf(b, "%s:\n - Author: %v\n - Content: %s\n", label, author, content)
}

func buildPrompt(m *messages.Message) string {
	var b strings.Builder

	if q := m.QuotedMessage; q != nil {
		author := q.Author.JID
		if q.IsFromMe {
			author = deref(m.Bot.Name)
		}
		writeMessage(&b, "Quoted Message", author, deref(q.Text))
		b.WriteString("\n")
	}

	writeMessage(&b, "Current Message", m.Author.JID, deref(m.Text))
	return b.String()
}

func Answer(m *messages.Message) error {
	c := chatter.New(m.Bot.DB.AIModel)

	answer, err := c.GenerateMessage(m.Ctx, deref(m.Bot.Name), m.Chat.Prefix, m.Chat.BotStyle, buildPrompt(m))
	if err != nil {
		return fmt.Errorf("llm: generate message: %w", err)
	}

	m.Reply(answer)
	return nil
}
