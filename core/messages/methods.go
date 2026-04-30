package messages

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func (m *Message) Reply(content string, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {
	if err := m.handleOptionalReaction(reaction...); err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("message content is empty")
	}

	mentions := utils.GenerateMentionFromText(content)
	context := m.buildReplyContext()
	context.MentionedJID = mentions.Mention
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(mentions.Text),
			ContextInfo: context,
			//ContextInfo: &waE2E.ContextInfo{
			//	QuotedMessage: &waE2E.Message{
			//		Conversation: proto.String(m.safeText()),
			//	},
			//	MentionedJID: mentions.Mention,
			//},
		},
	}

	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, msg)
	return &resp, err
}

func (m *Message) ReplyMedia(content []byte, caption string, _type MessageType, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {
	if err := m.handleOptionalReaction(reaction...); err != nil {
		return nil, err
	}
	return m.SendMediaMessage(content, caption, _type, m.RawEvent.Info.Chat, m.buildReplyContext())
}

func (m *Message) ReplySticker(content []byte, _type MessageType, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {
	if err := m.handleOptionalReaction(reaction...); err != nil {
		return nil, err
	}
	return m.SendStickerMessage(content, _type, m.RawEvent.Info.Chat, m.buildReplyContext())
}

func (m *Message) React(emoji emojis.Emoji) (*whatsmeow.SendResponse, error) {
	reactionMsg := m.Bot.Client.BuildReaction(m.RawEvent.Info.Chat, m.RawEvent.Info.Sender, m.RawEvent.Info.ID, string(emoji))
	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, reactionMsg)
	return &resp, err
}

func (m *Message) Delete() (*whatsmeow.SendResponse, error) {
	revokeMsg := m.Bot.Client.BuildRevoke(m.RawEvent.Info.Chat, m.RawEvent.Info.Sender, m.RawEvent.Info.ID)
	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, revokeMsg)
	return &resp, err
}

func (m *Message) Edit(content string) (*whatsmeow.SendResponse, error) {
	editMsg := m.Bot.Client.BuildEdit(m.RawEvent.Info.Chat, m.RawEvent.Info.ID, &waE2E.Message{
		Conversation: proto.String(content),
	})
	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, editMsg)
	return &resp, err
}

func (m *Message) IsFromGroup() bool {
	return m.RawEvent.Info.IsGroup
}

func (m *Message) SendMessage(msg *waE2E.Message, chatId types.JID) (*whatsmeow.SendResponse, error) {
	resp, err := m.Bot.Client.SendMessage(m.Ctx, chatId, msg)
	return &resp, err
}

func (m *Message) SendVideoMessage(caption string, video []byte, chatId types.JID, quoteMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	resp, err := m.uploadMedia(video, whatsmeow.MediaVideo)
	if err != nil {
		return nil, err
	}

	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String("video/mp4"),
			URL:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			ContextInfo:   getFirstContext(quoteMessage),
		},
	}

	return m.sendAndLog(chatId, msg, "video")
}

func (m *Message) SendImageMessage(caption string, image []byte, chatId types.JID, quotedMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	resp, err := m.uploadMedia(image, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String("image/jpeg"),
			URL:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			ContextInfo:   getFirstContext(quotedMessage),
		},
	}

	return m.sendAndLog(chatId, msg, "image")
}

func (m *Message) SendStickerMessage(media []byte, _type MessageType, chatId types.JID, quotedMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	var err error
	contentType := http.DetectContentType(media)

	if _type == ImageMessage && contentType != "image/webp" {
		if media, err = utils.ToWebp(media); err != nil {
			slog.Error("Error converting media to webp", "error", err)
			return nil, err
		}
		contentType = http.DetectContentType(media)
	}

	resp, err := m.uploadMedia(media, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	msg := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			Mimetype:      proto.String(contentType),
			URL:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			FileSHA256:    resp.FileSHA256,
			FileEncSHA256: resp.FileEncSHA256,
			MediaKey:      resp.MediaKey,
			FileLength:    &resp.FileLength,
			ContextInfo:   getFirstContext(quotedMessage),
			IsAnimated:    proto.Bool(_type == VideoMessage),
		},
	}

	return m.sendAndLog(chatId, msg, "sticker")
}

func (m *Message) SendAudioMessage(media []byte, chatId types.JID, quotedMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	resp, err := m.uploadMedia(media, whatsmeow.MediaAudio)
	if err != nil {
		return nil, err
	}

	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:               &resp.URL,
			DirectPath:        &resp.DirectPath,
			MediaKey:          resp.MediaKey,
			Mimetype:          proto.String("application/ogg"),
			FileEncSHA256:     resp.FileEncSHA256,
			FileSHA256:        resp.FileSHA256,
			FileLength:        proto.Uint64(resp.FileLength),
			ContextInfo:       getFirstContext(quotedMessage),
			StreamingSidecar:  resp.FileSHA256,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
		},
	}

	return m.sendAndLog(chatId, msg, "audio")
}

func (m *Message) SendMediaMessage(content []byte, caption string, _type MessageType, chatId types.JID, quoteMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	quote := getFirstContext(quoteMessage)

	switch _type {
	case ImageMessage:
		return m.SendImageMessage(caption, content, chatId, quote)
	case VideoMessage:
		return m.SendVideoMessage(caption, content, chatId, quote)
	case AudioMessage:
		return m.SendAudioMessage(content, chatId, quote)
	default:
		err := fmt.Errorf("unsupported media type: %v", _type)
		slog.Error(err.Error())
		return nil, err
	}
}

func (m *Message) HasValidMedia(ignoreSticker ...bool) bool {
	if len(ignoreSticker) > 0 && ignoreSticker[0] {
		return m.Type == ImageMessage || m.Type == VideoMessage
	}
	return m.Type == ImageMessage || m.Type == VideoMessage || m.Type == StickerMessage
}
