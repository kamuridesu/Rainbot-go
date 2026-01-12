package fun

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/sticker"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

var (
	ErrInvalidMediaType = errors.New("invalid media type for message")
)

func getMessageMediaBytes(m *messages.Message) ([]byte, error) {
	switch m.Type {
	case messages.ImageMessage:
		return m.Bot.Client.Download(m.Ctx, m.RawMessage.ImageMessage)
	case messages.VideoMessage:
		return m.Bot.Client.Download(m.Ctx, m.RawMessage.VideoMessage)
	default:
		return nil, ErrInvalidMediaType
	}
}

func newSticker(m *messages.Message, type_ sticker.StickerType) {

	slog.Info("Recv sticker request")
	var content []byte
	var err error
	contentType := m.Type

	m.React(emojis.Waiting)

	if m.HasValidMedia(true) {
		content, err = getMessageMediaBytes(m)
	} else if m.QuotedMessage != nil && m.QuotedMessage.HasValidMedia(true) {
		contentType = m.QuotedMessage.Type
		content, err = getMessageMediaBytes(m.QuotedMessage)
	} else {
		m.Reply("Nenhuma mídia encontrada na mensagem atual ou mencionada.", emojis.Fail)
		return
	}

	if err != nil {
		m.Reply(fmt.Sprintf("Houve um erro ao processar a mídia: %s", err), emojis.Fail)
		return
	}
	author := *m.Bot.Name
	pack := "bot"

	st := sticker.New(author, pack, content, type_)
	bytes, err := st.Convert()
	if err != nil {
		m.Reply(fmt.Sprintf("Falha ao converter para sticker: %s", err), emojis.Fail)
		return
	}

	f, err := sticker.CreateTempFile(bytes)
	if err == nil {
		slog.Info("Filename: " + f)
	}

	_, err = m.ReplySticker(bytes, contentType, emojis.Success)
	if err != nil {
		slog.Error(err.Error())
	}
}

func NewStickerOriginal(m *messages.Message) {
	newSticker(m, sticker.StickerOriginal)
}

func NewStickerSquash(m *messages.Message) {
	newSticker(m, sticker.StickerSquash)
}
