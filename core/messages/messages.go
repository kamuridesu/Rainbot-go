package messages

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type MessageType int

const (
	TextMessage MessageType = iota
	StickerMessage
	ImageMessage
	VideoMessage
	AudioMessage
	ReactionMessage
	Unknown
)

type Message struct {
	Ctx              context.Context
	Bot              *bot.Bot
	Args             *[]string
	RawEvent         *events.Message
	RawMessage       *waE2E.Message
	Text             *string
	Type             MessageType
	Command          *string
	Chat             *models.Chat
	Author           *models.Member
	Filters          []*models.Filter
	MentionedMembers []*models.Member
	QuotedMessage    *Message
}

type Handler struct {
	Ctx            context.Context
	Bot            *bot.Bot
	CommandHandler func(*Message)
	ChatHandler    func(*Message)
}

func NewHandler(ctx context.Context, commandHandler, chatHandler func(*Message)) *Handler {
	return &Handler{ctx, nil, commandHandler, chatHandler}
}

func (h *Handler) AttachBot(b *bot.Bot) {
	h.Bot = b
}

func deduplicateMentions(mentions []string) []string {
	keys := make(map[string]bool)
	tmp := []string{}

	for _, item := range mentions {
		if _, value := keys[item]; !value {
			keys[item] = true
			tmp = append(tmp, item)
		}
	}
	return tmp
}

func newMessage(ctx context.Context, bot *bot.Bot, v *events.Message, quotedMessage *waE2E.Message) (*Message, error) {
	var message Message
	message.Ctx = ctx
	message.Bot = bot
	message.RawEvent = v
	chat, err := bot.DB.Chat.GetOrCreateChat(v.Info.Chat.String())
	if err != nil {
		slog.Error(fmt.Sprintf("Error fetching chat info from db, err is: %v", err))
		return nil, err
	}
	authorJid := v.Info.Sender.ToNonAD().String()
	if !strings.Contains(authorJid, "@lid") {
		authorJid = v.Info.SenderAlt.ToNonAD().String()
		if !strings.Contains(authorJid, "@lid") {
			panic("User id is not lid: " + authorJid)
		}
	}
	member, err := bot.DB.Member.GetOrCreateMember(v.Info.Chat.String(), authorJid)
	if err != nil {
		slog.Error(fmt.Sprintf("Error fetching member info from db, err is: %v", err))
		return nil, err
	}
	filters, err := bot.DB.Filter.GetFilters(v.Info.Chat.String())
	if err != nil {
		slog.Error(fmt.Sprintf("Error fetching filter info from db, err is: %v", err))
		return nil, err
	}

	message.Chat = chat
	message.Author = member
	message.Filters = filters

	var messageText *string
	var mentionedJIDs []string

	rawMsg := v.Message
	if quotedMessage != nil {
		rawMsg = quotedMessage
	}
	message.RawMessage = rawMsg

	if rawMsg.Conversation != nil {
		message.Type = TextMessage
		messageText = rawMsg.Conversation
	} else if rawMsg.ExtendedTextMessage != nil {
		message.Type = TextMessage
		messageText = rawMsg.ExtendedTextMessage.Text
		if rawMsg.ExtendedTextMessage.ContextInfo != nil {
			mentionedJIDs = slices.Concat(mentionedJIDs, rawMsg.ExtendedTextMessage.ContextInfo.GetMentionedJID())
			if rawMsg.ExtendedTextMessage.ContextInfo.Participant != nil {
				mentionedJIDs = append(mentionedJIDs, *rawMsg.ExtendedTextMessage.ContextInfo.Participant)
			}
			if rawMsg.ExtendedTextMessage.ContextInfo.QuotedMessage != nil && quotedMessage == nil {
				message.QuotedMessage, err = newMessage(ctx, bot, v, rawMsg.ExtendedTextMessage.ContextInfo.QuotedMessage)
				if err != nil {
					slog.Error("error parsing quoted message")
					return nil, err
				}
			}
		}
	} else if rawMsg.ImageMessage != nil {
		messageText = rawMsg.ImageMessage.Caption
		message.Type = ImageMessage
		if rawMsg.ImageMessage.ContextInfo != nil {
			mentionedJIDs = slices.Concat(mentionedJIDs, rawMsg.ImageMessage.ContextInfo.GetMentionedJID())
			if participant := rawMsg.ImageMessage.ContextInfo.Participant; participant != nil {
				mentionedJIDs = append(mentionedJIDs, *participant)
			}
		}
	} else if rawMsg.VideoMessage != nil {
		message.Type = VideoMessage
		messageText = rawMsg.VideoMessage.Caption
		if rawMsg.VideoMessage.ContextInfo != nil {
			mentionedJIDs = slices.Concat(mentionedJIDs, rawMsg.VideoMessage.ContextInfo.GetMentionedJID())
			if participant := rawMsg.VideoMessage.ContextInfo.Participant; participant != nil {
				mentionedJIDs = append(mentionedJIDs, *participant)
			}
		}
	} else if rawMsg.StickerMessage != nil {
		message.Type = StickerMessage
	} else if rawMsg.AudioMessage != nil {
		message.Type = AudioMessage
	} else if rawMsg.ReactionMessage != nil {
		message.Type = ReactionMessage
	} else {
		message.Type = Unknown
	}

	mentionedJIDs = deduplicateMentions(mentionedJIDs)

	for _, jid := range mentionedJIDs {
		if strings.HasSuffix(jid, "whatsapp.net") {
			j := types.NewJID(strings.TrimSuffix(jid, "@s.whatsapp.net"), types.DefaultUserServer)

			j, err = bot.Client.Store.LIDs.GetLIDForPN(ctx, j)
			if err != nil {
				return nil, err
			}
			jid = j.ToNonAD().String()
		}
		m, err := bot.DB.Member.GetOrCreateMember(v.Info.Chat.String(), jid)
		if err != nil {
			return nil, err
		}
		message.MentionedMembers = append(message.MentionedMembers, m)
	}

	if messageText == nil {
		emptyMessage := ""
		messageText = &emptyMessage
	} else {
		noPrefix := strings.TrimPrefix(*messageText, chat.Prefix)
		parts := utils.ParseArgsFromMessage(noPrefix)
		if len(parts) > 0 {
			command := parts[0]
			args := parts[1:]
			message.Args = &args
			message.Command = &command
		}
	}

	message.Text = messageText

	return &message, nil
}

func (m *Message) HasValidMedia(ignoreSticker ...bool) bool {
	if len(ignoreSticker) > 0 && ignoreSticker[0] {
		return m.Type == ImageMessage || m.Type == VideoMessage
	}
	return m.Type == ImageMessage || m.Type == VideoMessage || m.Type == StickerMessage
}

func (h *Handler) Handle(event any) {
	if h.Bot == nil {
		panic(fmt.Errorf("no bot instance attached to handler"))
	}

	// slog.Info(fmt.Sprintf("Event: %v", event))
	switch v := event.(type) {
	case *events.Message:
		if v.Info.Timestamp.Before(h.Bot.StartTime) || v.Info.IsFromMe || v.Info.Chat.String() == "status@broadcast" {
			return
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// 		fmt.Println("=============== Received a message! =====================")
			// 		fmt.Println("Chat info: ")
			// 		fmt.Printf("Chat id: %s\n", v.Info.Chat.String())
			// 		fmt.Println("Sender Info: ")
			// 		fmt.Printf("User: %s\n", v.Info.Sender.SignalAddress().Name())
			// 		fmt.Printf("New ID: %s\n", v.Info.Sender.String())
			// 		fmt.Printf("Old ID: %s\n", v.Info.SenderAlt.String())
			// 		fmt.Printf("Chat AD: %s\n", v.Info.Sender.ADString())
			// 		fmt.Println("Message Info: ")
			//
			// 		if v.Message.ExtendedTextMessage != nil {
			// 			fmt.Printf("Mentions: %s\n", v.Message.ExtendedTextMessage.ContextInfo.GetMentionedJID())
			// 		}

			message, err := newMessage(ctx, h.Bot, v, nil)
			if err != nil {
				slog.Error("error parsing message, err is: " + err.Error())
				return
			}

			if message.Chat.IsBotEnabled == 0 {
				if !(strings.HasPrefix(*message.Text, message.Chat.Prefix) && (*message.Command == "setup")) {
					return
				}
			}

			if message.Text != nil {
				if strings.HasPrefix(*message.Text, message.Chat.Prefix) && !(message.Author.Silenced == 1) && message.Command != nil {
					slog.Info("Received message with prefix " + message.Chat.Prefix)
					h.CommandHandler(message)
				} else {
					if message.Chat.CountMessages == 1 && !(message.Author.Silenced == 1) {
						message.Author.Messages += 1
						h.Bot.DB.Member.Update(message.Author)
					}
					h.ChatHandler(message)
				}
			}
		}()
	case *events.GroupInfo:
		if len(v.Join) > 0 {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				chatInfo, err := h.Bot.DB.Chat.GetOrCreateChat(v.JID.String())
				if err != nil {
					slog.Error(err.Error())
					return
				}
				if chatInfo.WelcomeMessage == "" {
					return
				}
				var usrs []string
				for _, usr := range v.Join {
					usrs = append(usrs, utils.ParseLidToMention(usr.ToNonAD().String()))
				}
				msg := utils.GenerateMentionFromText(strings.Replace(chatInfo.WelcomeMessage, "@users", strings.Join(usrs, ", "), 1))
				h.Bot.Client.SendMessage(ctx, v.JID, &waE2E.Message{
					ExtendedTextMessage: &waE2E.ExtendedTextMessage{
						Text: &msg.Text,
						ContextInfo: &waE2E.ContextInfo{
							MentionedJID: msg.Mention,
						},
					},
				})
			}()
		}
	}
}

func (m *Message) SendMessage(msg *waE2E.Message, chatId types.JID) (*whatsmeow.SendResponse, error) {
	resp, err := m.Bot.Client.SendMessage(m.Ctx, chatId, msg)
	return &resp, err
}

func (m *Message) SendVideoMessage(caption string, video []byte, chatId types.JID, quoteMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	slog.Info("Uploading video")
	resp, err := m.Bot.Client.Upload(m.Ctx, video, whatsmeow.MediaVideo)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	var quote *waE2E.ContextInfo
	if len(quoteMessage) > 0 {
		quote = quoteMessage[0]
	}
	slog.Info("Video uploaded successfully")
	message := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String("video/mp4"),
			URL:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			ContextInfo:   quote,
		},
	}
	slog.Info("Sending media message")
	r, err := m.Bot.Client.SendMessage(m.Ctx, chatId, message)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("media sent successfully")
	return &r, err
}

func (m *Message) SendImageMessage(caption string, image []byte, chatId types.JID, quotedMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	slog.Info("Uploading image")
	resp, err := m.Bot.Client.Upload(m.Ctx, image, whatsmeow.MediaImage)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Image uploaded successfully")

	var quoted *waE2E.ContextInfo
	if len(quotedMessage) > 0 {
		quoted = quotedMessage[0]
	}
	message := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(caption),
			Mimetype:      proto.String("image/jpeg"),
			URL:           &resp.URL,
			DirectPath:    &resp.DirectPath,
			MediaKey:      resp.MediaKey,
			FileEncSHA256: resp.FileEncSHA256,
			FileSHA256:    resp.FileSHA256,
			FileLength:    &resp.FileLength,
			ContextInfo:   quoted,
		},
	}
	slog.Info("Sending media message")
	r, err := m.Bot.Client.SendMessage(m.Ctx, chatId, message)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("media sent successfully")
	return &r, err
}

func (m *Message) SendStickerMessage(media []byte, _type MessageType, chatId types.JID, quotedMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	type_ := whatsmeow.MediaImage
	var err error
	contentType := http.DetectContentType(media)
	switch _type {
	case VideoMessage:
	case ImageMessage:
		if contentType != "image/webp" {
			media, err = utils.ToWebp(media)
			contentType = http.DetectContentType(media)
			if err != nil {
				slog.Error("Error converting media: " + err.Error())
				return nil, err
			}
		}
	}
	resp, err := m.Bot.Client.Upload(m.Ctx, media, type_)
	if err != nil {
		slog.Error("Error uploading media: " + err.Error())
		return nil, err
	}
	var quoted *waE2E.ContextInfo
	if len(quotedMessage) > 0 {
		quoted = quotedMessage[0]
	}
	message := &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			Mimetype:      proto.String(contentType),
			URL:           proto.String(resp.URL),
			DirectPath:    proto.String(resp.DirectPath),
			FileSHA256:    resp.FileSHA256,
			FileEncSHA256: resp.FileEncSHA256,
			MediaKey:      resp.MediaKey,
			FileLength:    &resp.FileLength,
			ContextInfo:   quoted,
			IsAnimated:    proto.Bool(_type == VideoMessage),
		},
	}

	r, err := m.Bot.Client.SendMessage(m.Ctx, chatId, message)
	return &r, err
}

func (m *Message) ReplySticker(content []byte, type_ MessageType, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {
	r, err := m.SendStickerMessage(content, type_, m.RawEvent.Info.Chat, &waE2E.ContextInfo{
		StanzaID:    proto.String(m.RawEvent.Info.ID),
		Participant: proto.String(m.RawEvent.Info.Sender.String()),
		QuotedMessage: &waE2E.Message{
			Conversation: proto.String(*m.Text),
		},
	})
	if err != nil {
		slog.Error("Error sending sticker: " + err.Error())
		return nil, err
	}
	if len(reaction) != 0 {
		_, err := m.React(reaction[0])
		if err != nil {
			return nil, err
		}
	}
	return r, err
}

func (m *Message) SendAudioMessage(media []byte, chatId types.JID, quotedMessage ...*waE2E.ContextInfo) (*whatsmeow.SendResponse, error) {
	slog.Info("Uploading audio")
	resp, err := m.Bot.Client.Upload(m.Ctx, media, whatsmeow.MediaAudio)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("Audio uploaded successfully")

	var quoted *waE2E.ContextInfo
	if len(quotedMessage) > 0 {
		quoted = quotedMessage[0]
	}
	message := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:               proto.String(resp.URL),
			DirectPath:        proto.String(resp.DirectPath),
			MediaKey:          resp.MediaKey,
			Mimetype:          proto.String("application/ogg"),
			FileEncSHA256:     resp.FileEncSHA256,
			FileSHA256:        resp.FileSHA256,
			FileLength:        proto.Uint64(resp.FileLength),
			ContextInfo:       quoted,
			StreamingSidecar:  resp.FileSHA256,
			MediaKeyTimestamp: proto.Int64(time.Now().Unix()),
		},
	}
	slog.Info("Sending media message")
	r, err := m.Bot.Client.SendMessage(m.Ctx, chatId, message)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	slog.Info("media sent successfully")
	return &r, err

}

func (m *Message) SendMediaMessage(content []byte, caption string, _type MessageType, chatId types.JID, quoteMessage ...*waE2E.ContextInfo) {
	var err error
	var quote *waE2E.ContextInfo
	if len(quoteMessage) > 0 {
		quote = quoteMessage[0]
	}

	switch _type {
	case ImageMessage:
		_, err = m.SendImageMessage(caption, content, chatId, quote)
	case VideoMessage:
		_, err = m.SendVideoMessage(caption, content, chatId, quote)
	case AudioMessage:
		_, err = m.SendAudioMessage(content, chatId, quote)
	}
	if err != nil {
		slog.Error(err.Error())
	}
}

func (m *Message) ReplyMedia(content []byte, caption string, _type MessageType, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {
	react := func() error {
		if len(reaction) != 0 {
			_, err := m.React(reaction[0])
			if err != nil {
				return err
			}
		}
		return nil
	}

	quote := &waE2E.ContextInfo{
		StanzaID:    proto.String(m.RawEvent.Info.ID),
		Participant: proto.String(m.RawEvent.Info.Sender.String()),
		QuotedMessage: &waE2E.Message{
			Conversation: proto.String(*m.Text),
		},
	}

	m.SendMediaMessage(content, caption, _type, m.RawEvent.Info.Chat, quote)

	react()

	return nil, nil
}

func (m *Message) Reply(content string, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {

	if len(reaction) != 0 {
		_, err := m.React(reaction[0])
		if err != nil {
			return nil, err
		}
	}

	content = strings.TrimSpace(content)

	if content == "" {
		return nil, fmt.Errorf("err msg is empty")
	}

	mentions := utils.GenerateMentionFromText(content)
	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(mentions.Text),
			ContextInfo: &waE2E.ContextInfo{
				QuotedMessage: &waE2E.Message{
					Conversation: proto.String(*m.Text),
				},
				MentionedJID: mentions.Mention,
			},
		},
	}, whatsmeow.SendRequestExtra{})

	return &resp, err
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
