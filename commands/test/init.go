package test

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/quotly"
	"github.com/kamuridesu/rainbot-go/core/modules/sticker"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

func init() {
	commands.NewCommand("test", "", "test", nil, nil, false, false, false, func(m *messages.Message) {
		numMessages := 1
		hasR := false
		hasNumber := false

		if m.Args != nil {
			for _, arg := range *m.Args {
				if arg == "r" {
					hasR = true
				} else if val, err := strconv.Atoi(arg); err == nil {
					numMessages = val
					hasNumber = true
				}
			}
		}

		if hasNumber {

			hasR = false
			if numMessages > 5 {
				numMessages = 5
			}
			if numMessages < 1 {
				numMessages = 1
			}
		} else {

			numMessages = 1
		}

		slog.Info("Starting quote generation", "count", numMessages, "hasR", hasR, "hasNumber", hasNumber)

		primaryQuoteID := ""
		if ctxInfo := m.RawMessage.GetExtendedTextMessage().GetContextInfo(); ctxInfo != nil && ctxInfo.StanzaID != nil {
			primaryQuoteID = *ctxInfo.StanzaID
		}

		anchorMsg, err := m.Bot.DB.Message.GetMessage(primaryQuoteID)
		if err != nil || anchorMsg == nil {
			m.Reply("Could not find the quoted message.", emojis.Fail)
			return
		}

		history, err := m.Bot.DB.Message.GetMessageRange(m.Chat.ChatID, anchorMsg.CreatedAt, numMessages)
		if err != nil || len(history) == 0 {
			slog.Error("Failed to fetch message range", "err", err)
			return
		}

		resolveAuthor := func(msg *models.Message) (string, []byte) {
			pp, err := utils.DownloadIUserProfilePic(m.Ctx, msg.SenderJID, m.Bot)
			if err != nil || pp == nil {
				pp, _ = os.ReadFile("resources/images/default.png")
			}

			qJid, _ := types.ParseJID(msg.SenderJID)
			qContact, err := m.Bot.Client.Store.Contacts.GetContact(m.Ctx, qJid)
			name := ""
			if err == nil && qContact.Found {
				name = qContact.PushName
				if name == "" {
					name = qContact.FullName
				}
			}
			if name == "" {
				name = strings.Split(msg.SenderJID, "@")[0]
			}
			return name, pp
		}

		var quotlyMessages []quotly.QuotlyMessage

		for _, msg := range history {
			name, pp := resolveAuthor(msg)

			qMsg := quotly.QuotlyMessage{
				From: quotly.QuotlyUser{
					FirstName: name,
					Photo:     quotly.QuotlyUserPhoto{Url: utils.Encode64(pp, true)},
				},
				Text:   msg.MessageText,
				Avatar: true,
			}

			if hasR && msg.QuotedStanzaID != nil {
				parentMsg, err := m.Bot.DB.Message.GetMessage(*msg.QuotedStanzaID)
				if err == nil && parentMsg != nil {
					parentName, _ := resolveAuthor(parentMsg)

					qMsg.ReplyMessage = quotly.QuotlyReplyMessage{
						Name:   parentName,
						Text:   parentMsg.MessageText,
						ChatId: 1,
						From: quotly.QuotlyReplyUser{
							Id:   1,
							Name: parentName,
						},
					}
					slog.Info("Attached nested quote to payload based on 'r' flag", "parent", parentName)
				}
			}

			quotlyMessages = append(quotlyMessages, qMsg)
		}

		quotlyBody := quotly.DefaultTemplate
		quotlyBody.Messages = quotlyMessages

		img, err := quotly.Generate(m.Ctx, quotlyBody)
		if err != nil {
			slog.Error("Failed to generate quote", "err", err)
			m.Reply("Failed to generate quote.", emojis.Fail)
			return
		}

		packName := "Quotly"
		if m.Bot.Name != nil {
			packName = *m.Bot.Name
		}

		s := sticker.New(*m.Bot.Name, packName, img, sticker.StickerTransparent)
		stickerData, err := s.Convert()
		if err != nil {
			slog.Error("Failed to format sticker", "err", err)
			m.Reply("Failed to format the sticker.", emojis.Fail)
			return
		}

		m.ReplySticker(stickerData, messages.StickerMessage, emojis.Success)

	}, commands.HasQuotedMessage)
}
