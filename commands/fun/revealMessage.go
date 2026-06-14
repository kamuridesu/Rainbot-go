package fun

import (
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

func RevealMessage(m *messages.Message) {

	quoted := m.QuotedMessage
	type_ := quoted.Type
	var content []byte
	var err error

	switch quoted.Type {
	case messages.ImageMessage:
		content, err = m.Bot.Client.Download(m.Ctx, quoted.RawMessage.ImageMessage)
	case messages.VideoMessage:
		content, err = m.Bot.Client.Download(m.Ctx, quoted.RawMessage.VideoMessage)
	}

	if err != nil {
		m.Reply("Erro ao baixar midia: "+err.Error(), emojis.Fail)
		return
	}

	m.SendMediaMessage(content, "", type_, m.RawEvent.Info.Sender)
	m.React(emojis.Success)

}
