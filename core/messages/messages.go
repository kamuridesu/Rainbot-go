package messages

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

type MessageType int

const (
	TextMessage MessageType = iota
	StickerMessage
	ImageMessage
	VideoMessage
)

type Message struct {
	Ctx              context.Context
	Bot              *bot.Bot
	Args             *[]string
	RawEvent         *events.Message
	Text             *string
	Type             MessageType
	Command          *string
	Chat             *models.Chat
	Author           *models.Member
	Filters          []*models.Filter
	MentionedMembers []*models.Member
}

type Handler struct {
	Ctx           context.Context
	Bot           *bot.Bot
	CommandRunner func(*Message)
}

func NewHandler(ctx context.Context, commandRunner func(*Message)) *Handler {
	return &Handler{ctx, nil, commandRunner}
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

func newMessage(ctx context.Context, bot *bot.Bot, v *events.Message) (*Message, error) {
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
	if authorJid == "" {
		authorJid = v.Info.SenderAlt.ToNonAD().String()
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

	if v.Message.Conversation != nil {
		message.Type = TextMessage
		messageText = v.Message.Conversation
	} else if v.Message.ExtendedTextMessage != nil {
		message.Type = TextMessage
		messageText = v.Message.ExtendedTextMessage.Text
		mentionedJIDs = slices.Concat(mentionedJIDs, v.Message.ExtendedTextMessage.ContextInfo.GetMentionedJID())
		if v.Message.ExtendedTextMessage.ContextInfo.Participant != nil {
			mentionedJIDs = append(mentionedJIDs, *v.Message.ExtendedTextMessage.ContextInfo.Participant)
		}
	} else if v.Message.ImageMessage != nil {
		messageText = v.Message.ImageMessage.Caption
		message.Type = ImageMessage
		mentionedJIDs = slices.Concat(mentionedJIDs, v.Message.ImageMessage.ContextInfo.GetMentionedJID())
		if participant := v.Message.ImageMessage.ContextInfo.Participant; participant != nil {
			mentionedJIDs = append(mentionedJIDs, *participant)
		}
	} else if v.Message.VideoMessage != nil {
		message.Type = VideoMessage
		messageText = v.Message.VideoMessage.Caption
		mentionedJIDs = slices.Concat(mentionedJIDs, v.Message.VideoMessage.ContextInfo.GetMentionedJID())
		if participant := v.Message.VideoMessage.ContextInfo.Participant; participant != nil {
			mentionedJIDs = append(mentionedJIDs, *participant)
		}
	} else if v.Message.StickerMessage != nil {
		message.Type = StickerMessage
	}

	mentionedJIDs = deduplicateMentions(mentionedJIDs)

	for _, jid := range mentionedJIDs {
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
		parts := strings.Fields(noPrefix)
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

func (h *Handler) Handle(event any) {
	if h.Bot == nil {
		panic(fmt.Errorf("no bot instance attached to handler"))
	}
	switch v := event.(type) {
	case *events.Message:
		fmt.Println("=============== Received a message! =====================")
		fmt.Println("Chat info: ")
		fmt.Printf("Chat id: %s\n", v.Info.Chat.String())
		fmt.Println("Sender Info: ")
		fmt.Printf("User: %s\n", v.Info.Sender.SignalAddress().Name())
		fmt.Printf("New ID: %s\n", v.Info.Sender.String())
		fmt.Printf("Old ID: %s\n", v.Info.SenderAlt.String())
		fmt.Println("Message Info: ")

		if v.Message.ExtendedTextMessage != nil {
			fmt.Printf("Mentions: %s\n", v.Message.ExtendedTextMessage.ContextInfo.GetMentionedJID())
		}

		message, err := newMessage(h.Ctx, h.Bot, v)
		if err != nil {
			slog.Error("error parsing message, err is: " + err.Error())
			return
		}

		if message.Text != nil {
			if strings.HasPrefix(*message.Text, message.Chat.Prefix) {
				fmt.Println("Received message with prefix " + message.Chat.Prefix)
				h.CommandRunner(message)
			}
		}
		if v.Info.IsFromMe {
			// message.Reply(fmt.Sprintf("Você disse: %s", *message.Text))
		}
	}
}

func (m *Message) Reply(content string, reaction ...emojis.Emoji) (*whatsmeow.SendResponse, error) {
	if content == "" {
		return nil, fmt.Errorf("err msg is empty")
	}

	if len(reaction) != 0 {
		_, err := m.React(reaction[0])
		if err != nil {
			return nil, err
		}
	}

	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(content),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:    proto.String(m.RawEvent.Info.ID),
				Participant: proto.String(m.RawEvent.Info.Sender.String()),
				QuotedMessage: &waE2E.Message{
					Conversation: proto.String(*m.Text),
				},
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
