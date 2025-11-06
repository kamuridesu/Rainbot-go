package mute

import (
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/messages"
)

func DeleteIfMuted(m *messages.Message) {

	if m.Author.Silenced == 1 {

		msg := m.Bot.Client.BuildRevoke(m.RawEvent.Info.Chat, m.RawEvent.Info.Sender, m.RawEvent.Info.ID)
		_, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, msg)
		if err != nil {
			slog.Error("Error while deleting message: " + err.Error())
		}

	}

}
