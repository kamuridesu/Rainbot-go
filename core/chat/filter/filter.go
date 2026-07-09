package filter

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/storage"
)

var FilterCache sync.Map

type cachedRegex struct {
	re      *regexp.Regexp
	lastUse atomic.Int64
}

func getCompiledPattern(pattern string) (*regexp.Regexp, error) {
	if v, ok := FilterCache.Load(pattern); ok {
		entry := v.(*cachedRegex)
		entry.lastUse.Store(time.Now().Unix())
		return entry.re, nil
	}

	re, err := regexp.Compile(`(?i)(?:^|\s)` + regexp.QuoteMeta(pattern) + `(?:\s|$)`)
	if err != nil {
		return nil, err
	}

	entry := &cachedRegex{re: re}
	entry.lastUse.Store(time.Now().Unix())
	FilterCache.Store(pattern, entry)
	return re, nil
}

func StartCacheEviction(interval, ttl time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-ttl).Unix()
			FilterCache.Range(func(k, v any) bool {
				if v.(*cachedRegex).lastUse.Load() < cutoff {
					FilterCache.Delete(k)
				}
				return true
			})
		}
	}()
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
