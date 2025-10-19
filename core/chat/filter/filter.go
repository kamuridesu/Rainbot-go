package filter

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/storage"
)

func replyMedia(m *messages.Message, filter *models.Filter) error {
	file := storage.NewFile(filter.Response, storage.ModeReadOnly)
	bytes, err := file.Read()
	if err != nil {
		if errors.Is(err, storage.ErrNotExists) {
			slog.Error(fmt.Sprintf("File %s does not exists", file.Name))
			m.Reply(fmt.Sprintf("Arquivo do filter %s não foi econtrado, por favor, remova o filtro.", filter.Pattern), emojis.Fail)
			return nil
		}
		slog.Error(err.Error())
		return err
	}

	switch filter.Kind {
	case "image":
		_, err = m.ReplyMedia(bytes, "", messages.ImageMessage)
	case "video":
		_, err = m.ReplyMedia(bytes, "", messages.VideoMessage)
	case "sticker":
		_, err = m.ReplySticker(bytes, messages.ImageMessage)
	}

	return nil
}

func GetChatFilters(m *messages.Message) error {
	filters, err := m.Bot.DB.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		return err
	}

	for _, filter := range filters {
		if *m.Text == filter.Pattern {
			if filter.Kind == "text" {
				m.Reply(filter.Response)
				return nil
			} else {
				return replyMedia(m, filter)
			}
		}
	}
	return nil
}
