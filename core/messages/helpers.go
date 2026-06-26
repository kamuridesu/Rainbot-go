package messages

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

func fetchMessageContext(bot *bot.Bot, chatJID, authorJID string) (*models.Chat, *models.Member, []*models.Filter, error) {
	chat, err := bot.DB.Chat.GetOrCreateChat(chatJID)
	if err != nil {
		slog.Error("Error fetching chat info from db", "err", err)
		return nil, nil, nil, err
	}

	member, err := fetchMember(bot, chatJID, authorJID)
	if err != nil {
		slog.Error("Error fetching member info from db", "err", err)
		return nil, nil, nil, err
	}

	filters, err := bot.DB.Filter.GetFilters(chatJID)
	if err != nil {
		slog.Error("Error fetching filter info from db", "err", err)
		return nil, nil, nil, err
	}

	return chat, member, filters, nil
}

func fetchMember(bot *bot.Bot, chatJID, authorJID string) (*models.Member, error) {
	member, err := bot.DB.Member.GetOrCreateMember(chatJID, authorJID)
	if err != nil {
		slog.Error("Error fetching member info from db", "err", err)
		return nil, err
	}
	return member, nil
}

func getAuthorJID(v *events.Message) string {
	authorJid := v.Info.Sender.ToNonAD().String()
	if !strings.Contains(authorJid, "@lid") {
		authorJid = v.Info.SenderAlt.ToNonAD().String()
		if !strings.Contains(authorJid, "@lid") {
			panic("User id is not lid: " + authorJid)
		}
	}
	return authorJid
}

func parseMessageContent(rawMsg *waE2E.Message) (msgType MessageType, text *string, mentions []string, quoted *waE2E.Message, quotedStanzaID, quotedMessageAuthor string) {
	mentions = make([]string, 0)
	quotedStanzaID = ""
	quotedMessageAuthor = ""

	switch {
	case rawMsg.Conversation != nil:
		msgType = TextMessage
		text = rawMsg.Conversation

	case rawMsg.ExtendedTextMessage != nil:
		msgType = TextMessage
		text = rawMsg.ExtendedTextMessage.Text
		if ctxInfo := rawMsg.ExtendedTextMessage.ContextInfo; ctxInfo != nil {
			mentions = append(mentions, ctxInfo.GetMentionedJID()...)
			if ctxInfo.Participant != nil {
				mentions = append(mentions, *ctxInfo.Participant)
				quotedMessageAuthor = ctxInfo.GetParticipant()

			}
			quoted = ctxInfo.QuotedMessage
			if ctxInfo.StanzaID != nil {
				quotedStanzaID = *ctxInfo.StanzaID
			}
		}

	case rawMsg.ImageMessage != nil:
		msgType = ImageMessage
		text = rawMsg.ImageMessage.Caption
		if ctxInfo := rawMsg.ImageMessage.ContextInfo; ctxInfo != nil {
			mentions = append(mentions, ctxInfo.GetMentionedJID()...)
			if ctxInfo.Participant != nil {
				mentions = append(mentions, *ctxInfo.Participant)
			}
			if ctxInfo.StanzaID != nil {
				quotedStanzaID = *ctxInfo.StanzaID
			}
		}

	case rawMsg.VideoMessage != nil:
		msgType = VideoMessage
		text = rawMsg.VideoMessage.Caption
		if ctxInfo := rawMsg.VideoMessage.ContextInfo; ctxInfo != nil {
			mentions = append(mentions, ctxInfo.GetMentionedJID()...)
			if ctxInfo.Participant != nil {
				mentions = append(mentions, *ctxInfo.Participant)
			}
			if ctxInfo.StanzaID != nil {
				quotedStanzaID = *ctxInfo.StanzaID
			}
		}

	case rawMsg.StickerMessage != nil:
		msgType = StickerMessage
	case rawMsg.AudioMessage != nil:
		msgType = AudioMessage
	case rawMsg.ReactionMessage != nil:
		msgType = ReactionMessage
	default:
		msgType = Unknown
	}

	return msgType, text, mentions, quoted, quotedStanzaID, quotedMessageAuthor
}

func resolveMentions(ctx context.Context, bot *bot.Bot, chatJID string, rawMentions []string) ([]*models.Member, error) {
	if len(rawMentions) == 0 {
		return nil, nil
	}

	uniqueMentions := deduplicateMentions(rawMentions)
	var members []*models.Member

	for _, jid := range uniqueMentions {
		if strings.HasSuffix(jid, "whatsapp.net") {
			j := types.NewJID(strings.TrimSuffix(jid, "@s.whatsapp.net"), types.DefaultUserServer)

			resolvedJID, err := bot.Client.Store.LIDs.GetLIDForPN(ctx, j)
			if err != nil {
				return nil, err
			}
			jid = resolvedJID.ToNonAD().String()
		}

		m, err := bot.DB.Member.GetOrCreateMember(chatJID, jid)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

func parseCommandArgs(text, prefix string) (*string, *[]string) {
	noPrefix := strings.TrimPrefix(text, prefix)
	parts := utils.ParseArgsFromMessage(noPrefix)

	if len(parts) > 0 {
		cmd := parts[0]
		args := parts[1:]
		return &cmd, &args
	}

	return nil, nil
}

func (m *Message) uploadMedia(content []byte, mediaType whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
	slog.Info(fmt.Sprintf("Uploading %s", mediaType))
	resp, err := m.Bot.Client.Upload(m.Ctx, content, mediaType)
	if err != nil {
		slog.Error("Error uploading media", "type", mediaType, "error", err)
		return whatsmeow.UploadResponse{}, err
	}
	slog.Info(fmt.Sprintf("%s uploaded successfully", mediaType))
	return resp, nil
}

func (m *Message) sendAndLog(chatId types.JID, msg *waE2E.Message, mediaType string) (*whatsmeow.SendResponse, error) {
	slog.Info(fmt.Sprintf("Sending %s message", mediaType))
	resp, err := m.Bot.Client.SendMessage(m.Ctx, chatId, msg)
	if err != nil {
		slog.Error("Failed to send message", "type", mediaType, "error", err)
		return nil, err
	}
	slog.Info(fmt.Sprintf("%s sent successfully", mediaType))
	return &resp, nil
}

func (m *Message) buildReplyContext() *waE2E.ContextInfo {
	return &waE2E.ContextInfo{
		StanzaID:      proto.String(m.RawEvent.Info.ID),
		Participant:   proto.String(m.RawEvent.Info.Sender.ToNonAD().String()),
		QuotedMessage: m.RawMessage,
	}
}

func (m *Message) handleOptionalReaction(reactions ...emojis.Emoji) error {
	if len(reactions) > 0 {
		if _, err := m.React(reactions[0]); err != nil {
			slog.Error("Error sending reaction", "error", err)
			return err
		}
	}
	return nil
}

func (m *Message) safeText() string {
	if m.Text != nil {
		return *m.Text
	}
	return ""
}

func getFirstContext(ctxs []*waE2E.ContextInfo) *waE2E.ContextInfo {
	if len(ctxs) > 0 {
		return ctxs[0]
	}
	return nil
}
