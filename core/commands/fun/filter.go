package fun

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/storage"
)

func NewFilter(m *messages.Message) {

	quoted := m.QuotedMessage

	if quoted == nil {
		m.Reply("Preciso que uma mensagem seja mencionada.", emojis.Fail)
		return
	}

	if quoted.Type != messages.ImageMessage && quoted.Type != messages.VideoMessage && quoted.Type != messages.StickerMessage && quoted.Type != messages.TextMessage {
		m.Reply("Mensagem precisa de um corpo que seja imagem, texto, video ou figurinha.", emojis.Fail)
		return
	}

	filters, err := m.Bot.DB.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		m.Reply(fmt.Sprintf("Erro ao ler os filtros: %s", err), emojis.Fail)
		return
	}
	pattern := strings.Join(*m.Args, " ")
	_, err = regexp.Compile(pattern)
	if err != nil {
		m.Reply("Regex "+pattern+" invalido.", emojis.Fail)
		return
	}
	for _, filter := range filters {
		if filter.Pattern == pattern {
			m.Reply("Filtro já existe.", emojis.Fail)
			return
		}
	}

	var bytes []byte
	var filename string
	_type := "text"
	response := *quoted.Text

	if quoted.Type == messages.ImageMessage || quoted.Type == messages.VideoMessage || quoted.Type == messages.StickerMessage {
		_type = "media"
		switch quoted.Type {
		case messages.ImageMessage:
			_type = "image"
			bytes, err = m.Bot.Client.Download(m.Ctx, quoted.RawMessage.ImageMessage)
			filename = storage.RandomFilename("jpg")
		case messages.VideoMessage:
			_type = "video"
			bytes, err = m.Bot.Client.Download(m.Ctx, quoted.RawMessage.VideoMessage)
			filename = storage.RandomFilename("mp4")
		case messages.StickerMessage:
			_type = "sticker"
			bytes, err = m.Bot.Client.Download(m.Ctx, quoted.RawMessage.StickerMessage)
			filename = storage.RandomFilename("webp")
		}
		if err != nil {
			m.Reply(fmt.Sprintf("Erro ao baixar arquivo: %s", err), emojis.Fail)
			return
		}
		if filename != "" {
			file := storage.NewFile(filename)
			err := file.Write(bytes)
			if err != nil {
				m.Reply(fmt.Sprintf("Erro ao salvar arquivo: %s", err), emojis.Fail)
				return
			}
			response = filename
		}
	}

	filter := models.Filter{
		ChatID:   m.Chat.ChatID,
		Pattern:  pattern,
		Kind:     _type,
		Response: response,
	}

	err = m.Bot.DB.Filter.NewFilter(&filter)
	if err != nil {
		m.Reply(fmt.Sprintf("Falha ao salvar o filtro: %s", err), emojis.Fail)
		return
	}

	m.Reply("Filtro salvo com sucesso.", emojis.Success)
}

func ShowFilters(m *messages.Message) {
	filters, err := m.Bot.DB.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		m.Reply(fmt.Sprintf("Erro ao ler filters: %s", err), emojis.Fail)
		return
	}
	msg := ""
	for i, filter := range filters {
		msg += "- " + filter.Pattern
		if i < len(filters) {
			msg += "\n"
		}
	}

	m.Reply("// Filtros \\\\\n\n"+msg, emojis.Success)
}

func DeleteFilter(m *messages.Message) {
	pattern := strings.Join(*m.Args, " ")
	err := m.Bot.DB.Filter.Delete(m.Chat.ChatID, pattern)
	if err != nil {
		m.Reply(fmt.Sprintf("Erro ao deleter filtro: %s", err), emojis.Fail)
		return
	}
	m.React(emojis.Success)
}
