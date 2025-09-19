package messages

import (
	"context"
	"fmt"
	"strings"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Author struct {
	Jid  string
	Lid  string
	User string
}

type MessageType int

const (
	TextMessage MessageType = iota
	StickerMessage
	ImageMessage
	VideoMessage
)

type Message struct {
	Ctx      context.Context
	Bot      *bot.Bot
	Args     *[]string
	RawEvent *events.Message
	Text     *string
	Type     MessageType
	Command  *string
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

func newMessage(ctx context.Context, bot *bot.Bot, v *events.Message) *Message {
	var message Message
	message.Ctx = ctx
	message.Bot = bot
	message.RawEvent = v

	var messageText *string

	if v.Message.Conversation != nil {
		messageText = v.Message.Conversation
	}

	if messageText == nil {
		emptyMessage := ""
		messageText = &emptyMessage
	} else {
		noPrefix := strings.TrimPrefix(*messageText, *bot.Prefix)
		parts := strings.Split(noPrefix, " ")
		command := parts[0]
		args := parts[1:]
		message.Args = &args
		message.Command = &command

	}

	message.Text = messageText

	return &message
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

		message := newMessage(h.Ctx, h.Bot, v)

		if message.Text != nil {
			if strings.HasPrefix(*message.Text, *h.Bot.Prefix) {
				fmt.Println("Received message with prefix " + *h.Bot.Prefix)
				h.CommandRunner(message)
			}
		}
		fmt.Printf("Message: %s\n", *message.Text)
		if v.Info.IsFromMe {
			// message.Reply(fmt.Sprintf("Você disse: %s", *message.Text))
		}
	}
}

func (m *Message) Reply(content string) (*types.MessageID, error) {
	if content == "" {
		return nil, fmt.Errorf("err msg is empty")
	}
	resp, err := m.Bot.Client.SendMessage(m.Ctx, m.RawEvent.Info.Chat, &waE2E.Message{
		Conversation: &content,
	}, whatsmeow.SendRequestExtra{})

	return &resp.ID, err
}

func (m *Message) React() {}

func (m *Message) Delete() {}

func (m *Message) Edit() {}
