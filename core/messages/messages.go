package messages

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
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
	rootCtx          context.Context
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
	IsFromMe         bool
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

func newMessage(rootCtx, ctx context.Context, bot *bot.Bot, v *events.Message, quotedMessage *waE2E.Message) (*Message, error) {
	chatJID := v.Info.Chat.String()

	chat, member, filters, err := fetchMessageContext(bot, chatJID, getAuthorJID(v))
	if err != nil {
		return nil, err
	}

	rawMsg := v.Message
	if quotedMessage != nil {
		rawMsg = quotedMessage
	}

	msgType, text, rawMentions, nextQuotedMsg, _, quotedMessageAuthor := parseMessageContent(rawMsg)

	message := &Message{
		Ctx:        ctx,
		rootCtx:    rootCtx,
		Bot:        bot,
		RawEvent:   v,
		RawMessage: rawMsg,
		Chat:       chat,
		Author:     member,
		Filters:    filters,
		Type:       msgType,
		IsFromMe:   quotedMessage == nil && v.Info.IsFromMe,
	}

	if nextQuotedMsg != nil && quotedMessage == nil {
		message.QuotedMessage, err = newMessage(rootCtx, ctx, bot, v, nextQuotedMsg)
		if err != nil {
			slog.Error("Error parsing quoted message", "err", err)
			return nil, err
		}

		if quotedMessageAuthor != "" {
			if jid, parseErr := types.ParseJID(quotedMessageAuthor); parseErr == nil {
				message.QuotedMessage.IsFromMe = bot.IsSelf(jid)
			}

			quotedMember, fetchErr := fetchMember(bot, chatJID, quotedMessageAuthor)
			if fetchErr == nil {
				message.QuotedMessage.Author = quotedMember
			} else {
				slog.Warn("Could not fetch quoted message author", "jid", quotedMessageAuthor, "err", fetchErr)
			}
		}
	}

	message.MentionedMembers, err = resolveMentions(ctx, bot, chatJID, rawMentions)
	if err != nil {
		return nil, err
	}

	if text == nil {
		emptyStr := ""
		text = &emptyStr
	}
	message.Text = text

	if *text != "" {
		message.Command, message.Args = parseCommandArgs(*text, chat.Prefix)
	}

	return message, nil
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
			ctx, cancel := context.WithTimeout(h.Ctx, 60*time.Second)
			defer cancel()

			message, err := newMessage(h.Ctx, ctx, h.Bot, v, nil)
			if err != nil {
				slog.Error("error parsing message, err is: " + err.Error())
				return
			}

			if message.Chat.LeftAt != nil {
				if err := h.Bot.DB.Chat.ClearLeft(message.Chat.ChatID); err != nil {
					slog.Error("Failed to clear chat left mark", "chat", message.Chat.ChatID, "err", err)
				}
			}

			if message.Author.LeftAt != nil {
				if err := h.Bot.DB.Member.ClearLeft(message.Chat.ChatID, message.Author.JID); err != nil {
					slog.Error("Failed to clear member left mark", "chat", message.Chat.ChatID, "member", message.Author.JID, "err", err)
				}
			}

			timestamp := message.RawEvent.Info.Timestamp
			if timestamp.IsZero() {
				timestamp = time.Now()
			}

			var quotedID *string
			if raw := message.RawMessage; raw != nil {
				if ext := raw.GetExtendedTextMessage(); ext != nil && ext.ContextInfo != nil {
					quotedID = ext.ContextInfo.StanzaID
				} else if img := raw.GetImageMessage(); img != nil && img.ContextInfo != nil {
					quotedID = img.ContextInfo.StanzaID
				} else if vid := raw.GetVideoMessage(); vid != nil && vid.ContextInfo != nil {
					quotedID = vid.ContextInfo.StanzaID
				} else if doc := raw.GetDocumentMessage(); doc != nil && doc.ContextInfo != nil {
					quotedID = doc.ContextInfo.StanzaID
				} else if aud := raw.GetAudioMessage(); aud != nil && aud.ContextInfo != nil {
					quotedID = aud.ContextInfo.StanzaID
				} else if stk := raw.GetStickerMessage(); stk != nil && stk.ContextInfo != nil {
					quotedID = stk.ContextInfo.StanzaID
				}
			}

			if quotedID != nil && *quotedID == message.RawEvent.Info.ID {
				slog.Warn("Prevented message from quoting itself", "stanzaId", *quotedID)
				quotedID = nil
			}

			dbMsg := &models.Message{
				StanzaID:       message.RawEvent.Info.ID,
				ChatID:         message.Chat.ChatID,
				SenderJID:      message.Author.JID,
				MessageText:    *message.Text,
				QuotedStanzaID: quotedID,
				CreatedAt:      timestamp,
			}

			if err := h.Bot.DB.Message.SaveMessage(dbMsg); err != nil {
				slog.Error("Failed to save message to database", "err", err)
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
	case *events.JoinedGroup:
		if err := h.Bot.DB.Chat.ClearLeft(v.JID.String()); err != nil {
			slog.Error("Failed to clear chat left mark", "chat", v.JID.String(), "err", err)
		}
	case *events.GroupInfo:
		if v.Delete != nil || h.Bot.ContainsSelf(v.Leave) {
			slog.Info("Bot left chat, scheduling cleanup", "chat", v.JID.String())
			if err := h.Bot.DB.Chat.MarkLeft(v.JID.String()); err != nil {
				slog.Error("Failed to mark chat as left", "chat", v.JID.String(), "err", err)
			}
			return
		}
		if h.Bot.ContainsSelf(v.Join) {
			if err := h.Bot.DB.Chat.ClearLeft(v.JID.String()); err != nil {
				slog.Error("Failed to clear chat left mark", "chat", v.JID.String(), "err", err)
			}
		}
		if len(v.Leave) > 0 || len(v.Join) > 0 {
			go h.updateMembersLeft(v)
		}
		if len(v.Join) > 0 {
			go func() {
				ctx, cancel := context.WithTimeout(h.Ctx, 15*time.Second)
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

func (h *Handler) updateMembersLeft(v *events.GroupInfo) {
	ctx, cancel := context.WithTimeout(h.Ctx, 15*time.Second)
	defer cancel()

	chatJID := v.JID.String()
	for _, jid := range v.Leave {
		if lid := h.Bot.ToLID(ctx, jid); lid != "" {
			if err := h.Bot.DB.Member.MarkLeft(chatJID, lid); err != nil {
				slog.Error("Failed to mark member as left", "chat", chatJID, "member", lid, "err", err)
			}
		}
	}
	for _, jid := range v.Join {
		if lid := h.Bot.ToLID(ctx, jid); lid != "" {
			if err := h.Bot.DB.Member.ClearLeft(chatJID, lid); err != nil {
				slog.Error("Failed to clear member left mark", "chat", chatJID, "member", lid, "err", err)
			}
		}
	}
}
