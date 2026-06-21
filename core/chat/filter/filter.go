package filter

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sync"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/storage"
)

var FilterCache sync.Map

func getCompiledPattern(pattern string) (*regexp.Regexp, error) {
	if v, ok := FilterCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(pattern) + `\b`)
	if err != nil {
		return nil, err
	}
	FilterCache.Store(pattern, re)
	return re, nil
}

func matchesPattern(text, pattern string) bool {
	re, err := getCompiledPattern(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(text)
}

func replyMedia(m *messages.Message, filter *models.Filter) error {
	file := storage.NewFile(filter.Response, storage.ModeReadOnly)
	bytes, err := file.Read(m.Ctx)
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

	return err
}

func GetChatFilters(m *messages.Message) error {
	filters, err := m.Bot.DB.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		return err
	}
	for _, filter := range filters {
		if matchesPattern(*m.Text, filter.Pattern) {
			if filter.Kind == "text" {
				m.Reply(filter.Response)
				return nil
			}
			return replyMedia(m, filter)
		}
	}
	return nil
}
