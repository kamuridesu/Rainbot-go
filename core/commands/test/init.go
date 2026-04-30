package test

import (
	"errors"
	"log/slog"
	"os"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/quotly"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow/types"
)

func init() {
	commands.NewCommand("test", "", "test", nil, nil, false, false, false, func(m *messages.Message) {
		pp, err := utils.DownloadIUserProfilePic(m.Ctx, m.MentionedMembers[0].JID, m.Bot)
		if err != nil || pp == nil {
			if err != nil {
				slog.Warn("Could not retrieve user profile picture: " + err.Error())
			}
			pp, err = os.ReadFile("resources/images/default.png")
			if err != nil {
				slog.Error("Could not read default profile image: " + err.Error())
				return
			}
		}

		mentionedJid, err := types.ParseJID(m.MentionedMembers[0].JID)
		if err != nil {
			panic(err)
		}

		contact, err := m.Bot.Client.Store.Contacts.GetContact(m.Ctx, mentionedJid)
		if err != nil {
			panic(err)
		}
		if !contact.Found {
			panic(errors.New("no mentioned user found"))
		}

		username := contact.PushName
		if username == "" {
			username = contact.FullName
		}

		slog.Info(username)

		var replyMessage quotly.QuotlyReplyMessage
		if m.QuotedMessage.QuotedMessage != nil {
			qmsg := m.QuotedMessage.QuotedMessage
			jid, err := types.ParseJID(qmsg.Author.JID)
			if err != nil {
				panic(err)
			}
			contact, err := m.Bot.Client.Store.Contacts.GetContact(m.Ctx, jid)
			if err != nil {
				panic(err)
			}
			if !contact.Found {
				panic(errors.New("no mentioned user found"))
			}

			usernameMentioned := contact.PushName
			if usernameMentioned == "" {
				usernameMentioned = contact.FullName
			}
			replyMessage = quotly.QuotlyReplyMessage{
				Name:   usernameMentioned,
				Text:   *qmsg.Text,
				ChatId: 1,
				From: quotly.QuotlyReplyUser{
					Id:   1,
					Name: usernameMentioned,
				},
			}
		}

		quotlyBody := quotly.DefaultTemplate
		quotlyBody.Messages = []quotly.QuotlyMessage{
			{
				From: quotly.QuotlyUser{
					Id:        1,
					FirstName: username,
					LastName:  "",
					Username:  "test",
					Photo:     quotly.QuotlyUserPhoto{Url: utils.Encode64(pp, true)},
				},
				Text:         *m.QuotedMessage.Text,
				Avatar:       true,
				ReplyMessage: replyMessage,
			},
		}

		img, err := quotly.Generate(m.Ctx, quotlyBody)

		if err != nil {
			panic(err)
		}

		m.ReplySticker(img, messages.ImageMessage, emojis.Success)
	}, commands.HasQuotedMessage)
}
