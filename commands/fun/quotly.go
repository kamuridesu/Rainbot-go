package fun

import (
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/quotly"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

func HandleQuoteCommand(msg *messages.Message) {

	var targetStanzaID string
	if msg.QuotedMessage != nil {
		targetStanzaID = msg.QuotedMessage.RawEvent.Info.ID
	} else if msg.RawMessage.GetExtendedTextMessage().GetContextInfo().GetStanzaID() != "" {
		targetStanzaID = msg.RawMessage.GetExtendedTextMessage().GetContextInfo().GetStanzaID()
	} else {
		msg.Reply("Please reply to a message to quote it.")
		return
	}

	dbMsg, err := msg.Bot.DB.Message.GetMessage(targetStanzaID)
	if err != nil || dbMsg == nil {
		msg.Reply("Could not find the quoted message in the database. 🔍")
		return
	}

	var ppBase64 string
	ppBytes, err := utils.DownloadIUserProfilePic(msg.Ctx, dbMsg.SenderJID, msg.Bot)
	if err == nil && len(ppBytes) > 0 {
		ppBase64 = utils.Encode64(ppBytes, true)
	}

	displayName := strings.Split(dbMsg.SenderJID, "@")[0]

	qMsg := quotly.QuotlyMessage{
		From: quotly.QuotlyUser{
			Id:        1,
			FirstName: displayName,
			Photo:     quotly.QuotlyUserPhoto{Url: ppBase64},
		},
		Text:   dbMsg.MessageText,
		Avatar: true,
	}

	if dbMsg.QuotedStanzaID != nil {
		replyDbMsg, err := msg.Bot.DB.Message.GetMessage(*dbMsg.QuotedStanzaID)
		if err == nil && replyDbMsg != nil {
			replyDisplayName := strings.Split(replyDbMsg.SenderJID, "@")[0]
			qMsg.ReplyMessage = quotly.QuotlyReplyMessage{
				Name:   replyDisplayName,
				Text:   replyDbMsg.MessageText,
				ChatId: 0,
			}
		}
	}

	reqBody := quotly.DefaultTemplate
	reqBody.Messages = []quotly.QuotlyMessage{qMsg}

	imgBytes, err := quotly.Generate(msg.Ctx, reqBody)
	if err != nil {
		msg.Reply("Failed to generate quote image: "+err.Error(), emojis.Fail)
		return
	}

	msg.ReplySticker(imgBytes, messages.ImageMessage)
}
