package fun

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"time"

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

func NewAudioSticker(m *messages.Message) {
	var video *messages.Message
	switch {
	case m.Type == messages.VideoMessage:
		video = m
	case m.QuotedMessage != nil && m.QuotedMessage.Type == messages.VideoMessage:
		video = m.QuotedMessage
	default:
		m.Reply("Envie ou mencione um vídeo com som.", emojis.Fail)
		return
	}

	m.React(emojis.Waiting)

	content, err := getMessageMediaBytes(video)
	if err != nil {
		m.Reply(fmt.Sprintf("Houve um erro ao processar a mídia: %s", err), emojis.Fail)
		return
	}

	var maxLength time.Duration
	if m.Args != nil && len(*m.Args) > 0 {
		secs, err := strconv.Atoi((*m.Args)[0])
		if err != nil || secs < 1 || secs > sticker.AudioStickerMaxSeconds {
			m.Reply(fmt.Sprintf("Duração inválida, use um número de 1 a %d.", sticker.AudioStickerMaxSeconds), emojis.Fail)
			return
		}
		maxLength = time.Duration(secs) * time.Second
	}

	audio, length, err := sticker.ExtractAudio(content, maxLength)
	if errors.Is(err, sticker.ErrNoAudio) {
		m.Reply("Esse vídeo não tem som.", emojis.Fail)
		return
	}
	if err != nil {
		m.Reply(fmt.Sprintf("Falha ao extrair o áudio: %s", err), emojis.Fail)
		return
	}

	st := sticker.New(*m.Bot.Name, "bot", content, sticker.StickerSquash)
	st.Length = length
	webp, err := st.Convert()
	if err != nil {
		m.Reply(fmt.Sprintf("Falha ao converter para sticker: %s", err), emojis.Fail)
		return
	}

	seconds := uint32(max(1, math.Round(length.Seconds())))
	if _, err := m.ReplyAudioSticker(webp, audio, seconds, emojis.Success); err != nil {
		slog.Error("Failed to send audio sticker", "err", err)
	}
}
