package fun

import (
	"fmt"
	"regexp"
	"strings"

	cFilter "github.com/kamuridesu/rainbot-go/core/chat/filter"
	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
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
			err := file.Write(m.Ctx, bytes)
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

	if len(filters) == 0 {
		m.Reply("Nenhum filtro encontrado no chat.", emojis.Fail)
		return
	}

	var msg strings.Builder
	for i, filter := range filters {
		msg.WriteString("- ")
		msg.WriteString(filter.Pattern)
		if i < len(filters) {
			msg.WriteString("\n")
		}
	}

	m.Reply("// Filtros \\\\\n\n"+msg.String(), emojis.Success)
}

func DeleteFilter(m *messages.Message) {
	filters, err := m.Bot.DB.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		return
	}

	filterPattern := strings.Join(*m.Args, " ")

	var foundFilter *models.Filter
	for _, filter := range filters {
		if filterPattern == filter.Pattern {
			foundFilter = filter
		}
	}

	if foundFilter == nil {
		m.Reply("No filter with pattern "+filterPattern+" found", emojis.Fail)
		return
	}

	err = m.Bot.DB.Filter.Delete(m.Chat.ChatID, foundFilter.Pattern)
	if err != nil {
		m.Reply(fmt.Sprintf("Erro ao deleter filtro: %s", err), emojis.Fail)
		return
	}

	if foundFilter.Kind != "text" {
		file := storage.NewFile(foundFilter.Response)
		if e := file.Delete(m.Ctx); e != nil {
			m.Reply("Erro ao deletar arquivo do filtro", emojis.Fail)
		}
	}

	cFilter.FilterCache.Delete(foundFilter.Pattern)
	m.React(emojis.Success)
}
