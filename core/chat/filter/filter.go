package filter

import (
	"errors"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/sticker"
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

const (
	KindText         = "text"
	KindImage        = "image"
	KindVideo        = "video"
	KindSticker      = "sticker"
	KindAudioSticker = "audiosticker"
)

func AudioFilename(webpFilename string) string {
	return strings.TrimSuffix(webpFilename, path.Ext(webpFilename)) + ".ogg"
}

var errMissingFile = errors.New("filter file missing")

func readFilterFile(m *messages.Message, filter *models.Filter, name string) ([]byte, error) {
	file := storage.NewFile(name, storage.ModeReadOnly)
	bytes, err := file.Read(m.Ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNotExists) {
			slog.Error(fmt.Sprintf("File %s does not exists", file.Name))
			m.Reply(fmt.Sprintf("Arquivo do filter %s não foi econtrado, por favor, remova o filtro.", filter.Pattern), emojis.Fail)
			return nil, errMissingFile
		}
		slog.Error(err.Error())
		return nil, err
	}
	return bytes, nil
}

func replyAudioSticker(m *messages.Message, filter *models.Filter, webp []byte) error {
	audio, err := readFilterFile(m, filter, AudioFilename(filter.Response))
	if err != nil {
		return err
	}
	length, err := sticker.OggOpusDuration(audio)
	if err != nil {
		return fmt.Errorf("filter %s audio: %w", filter.Pattern, err)
	}
	_, err = m.ReplyAudioSticker(webp, audio, sticker.DurationSeconds(length))
	return err
}

func replyMedia(m *messages.Message, filter *models.Filter) error {
	bytes, err := readFilterFile(m, filter, filter.Response)
	if errors.Is(err, errMissingFile) {
		return nil
	}
	if err != nil {
		return err
	}

	switch filter.Kind {
	case KindImage:
		_, err = m.ReplyMedia(bytes, "", messages.ImageMessage)
	case KindVideo:
		_, err = m.ReplyMedia(bytes, "", messages.VideoMessage)
	case KindSticker:
		_, err = m.ReplySticker(bytes, messages.ImageMessage)
	case KindAudioSticker:
		err = replyAudioSticker(m, filter, bytes)
		if errors.Is(err, errMissingFile) {
			return nil
		}
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
			if filter.Kind == KindText {
				m.Reply(filter.Response)
				return nil
			}
			return replyMedia(m, filter)
		}
	}
	return nil
}
