package profanity

import (
	"fmt"
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/profanity"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func CheckForWord(m *messages.Message) bool {
	if m.Chat.ProfanityFilterEnabled == 0 {
		return false
	}

	dErr := profanity.HasObsceneWord(*m.Text)
	if dErr == nil {
		dErr = profanity.CheckCustomWord(m.Chat, *m.Text)
	}

	if dErr != nil {
		msg := m.Bot.Client.BuildRevoke(m.RawEvent.Info.Chat, m.RawEvent.Info.Sender, m.RawEvent.Info.ID)
		_, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, msg)
		if err != nil {
			slog.Error("Error while deleting message: " + err.Error())
		}
		m.Author.Warns += 1
		m.Bot.DB.Member.Update(m.Author)
		message := utils.GenerateMentionFromText(
			fmt.Sprintf(
				"%s\n\n%s recebeu 1 aviso. Mais %d e será banido.",
				dErr.Error(),
				utils.ParseLidToMention(m.Author.JID),
				m.Chat.WarnBanThreshold-m.Author.Warns,
			),
		)
		msg = &waE2E.Message{
			ExtendedTextMessage: &waE2E.ExtendedTextMessage{
				Text: &message.Text,
				ContextInfo: &waE2E.ContextInfo{
					MentionedJID: message.Mention,
				},
			},
		}
		_, err = m.SendMessage(msg, m.RawEvent.Info.Chat)
		if err != nil {
			slog.Error("Error while sending report message: " + err.Error())
		}
		return true
	}
	return false
}
