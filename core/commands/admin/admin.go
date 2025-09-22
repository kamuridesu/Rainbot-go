package admin

import (
	"fmt"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func banUser(m *messages.Message, users []*models.Member) (string, error) {
	var JIDs []types.JID

	message := ""

	for _, member := range users {
		jid, err := types.ParseJID(member.JID)
		if err != nil {
			return "", fmt.Errorf("Erro ao processar membros, erro é: %s", err.Error())
		}
		JIDs = append(JIDs, jid)
		message += "User " + member.JID + " removido!\n"
	}

	message = strings.TrimSpace(message)

	_, err := m.Bot.Client.UpdateGroupParticipants(m.RawEvent.Info.Chat, JIDs, whatsmeow.ParticipantChangeRemove)
	if err != nil {
		return "", err
	}

	_, err = m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		return "", err
	}

	return message, nil

}

func BanUser(m *messages.Message) {

	_, err := banUser(m, m.MentionedMembers)
	if err != nil {
		m.Reply(err.Error(), emojis.Fail)
		return
	}

	m.React(emojis.Success)

}

func WarnUser(m *messages.Message) {

	message := ""

	var toBeBanned []*models.Member

	for _, user := range m.MentionedMembers {
		user.Warns++
		m.Bot.DB.Member.Update(user)
		if user.Warns >= m.Chat.WarnBanThreshold {
			toBeBanned = append(toBeBanned, user)
			continue
		}
		message += fmt.Sprintf("User %s tem %d avisos, mais %d e será banido!\n", user.JID, user.Warns, m.Chat.WarnBanThreshold-user.Warns)
	}

	if len(toBeBanned) > 0 {
		msg, err := banUser(m, toBeBanned)
		if err != nil {
			m.Reply(err.Error(), emojis.Fail)
			return
		}
		message += "\n" + msg
	}

	message = strings.TrimSpace(message)

	m.Reply(message, emojis.Success)

}

func RemoveUserWarn(m *messages.Message) {

	message := ""

	for _, user := range m.MentionedMembers {
		if user.Warns < 1 {
			continue
		}
		user.Warns--
		m.Bot.DB.Member.Update(user)
		message += fmt.Sprintf("Aviso removido, agora %s tem %d avisos.", user.JID, user.Warns)
	}

	m.Reply(message, emojis.Success)

}

func MentionMembers(m *messages.Message) {

	message := ""

	if len(*m.Args) > 0 {
		message = strings.Join(*m.Args, " ")
	}

	if message == "" && m.RawEvent.Message.ExtendedTextMessage != nil && m.RawEvent.Message.ExtendedTextMessage.ContextInfo.QuotedMessage != nil {
		message = *m.RawEvent.Message.ExtendedTextMessage.ContextInfo.QuotedMessage.Conversation
	}

	if message == "" {
		m.Reply("Nenhuma mensagem recebida como argumento ou mencionada.", emojis.Fail)
		return
	}

	group, err := m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		m.Reply(fmt.Sprintf("Houve um erro ao processar dados do grupo: %s\n", err.Error()))
		return
	}

	var jids []string

	for _, member := range group.Participants {
		jids = append(jids, member.JID.String())
	}

	_, err = m.SendMessage(&waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &message,
			ContextInfo: &waE2E.ContextInfo{
				MentionedJID: jids,
			},
		},
	})
	if err != nil {
		m.Reply(fmt.Sprintf("Falha ao enviar mensagem: %s", err.Error()), emojis.Fail)
		return
	}

	m.React(emojis.Success)

}
