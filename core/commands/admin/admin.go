package admin

import (
	"fmt"
	"slices"
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
		if user.Warns >= m.Chat.WarnBanThreshold {
			toBeBanned = append(toBeBanned, user)
			user.Warns = 0
		} else {
			message += fmt.Sprintf("User %s tem %d avisos, mais %d e será banido!\n", user.JID, user.Warns, m.Chat.WarnBanThreshold-user.Warns)

		}
		m.Bot.DB.Member.Update(user)
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
	}, m.RawEvent.Info.Chat)
	if err != nil {
		m.Reply(fmt.Sprintf("Falha ao enviar mensagem: %s", err.Error()), emojis.Fail)
		return
	}
	m.React(emojis.Success)
}

func changeUserAdminStatus(m *messages.Message, demote ...bool) error {

	var usersToPromote []types.JID

	for _, user := range m.MentionedMembers {
		jid, err := types.ParseJID(user.JID)
		if err != nil {
			return err
		}
		usersToPromote = append(usersToPromote, jid)
	}

	action := whatsmeow.ParticipantChangePromote
	if demote != nil && demote[0] {
		action = whatsmeow.ParticipantChangeDemote
	}
	m.Bot.Client.UpdateGroupParticipants(m.RawEvent.Info.Chat, usersToPromote, action)

	return nil

}

func MessagesPerMember(m *messages.Message) {
	members, err := m.Bot.DB.Member.GetByChat(m.Chat.ChatID)
	if err != nil {
		m.Reply("Erro: "+err.Error(), emojis.Fail)
		return
	}

	slices.SortStableFunc(members, func(a, b *models.Member) int {
		return b.Messages - a.Messages
	})

	msg := "Total de mensagens por membros: \n\n"

	for i, member := range members {
		if member.Messages == 0 {
			continue
		}
		msg += fmt.Sprintf(`- %s: %d`, member.JID, member.Messages)
		if i < len(members)-1 {
			msg += "\n"
		}
	}

	m.Reply(msg, emojis.Success)
}

func PurgeMessages(m *messages.Message) {
	members, err := m.Bot.DB.Member.GetByChat(m.Chat.ChatID)
	if err != nil {
		m.Reply("Erro: "+err.Error(), emojis.Fail)
		return
	}

	for _, member := range members {
		member.Messages = 0
		m.Bot.DB.Member.Update(member)
	}

	m.React(emojis.Success)
}

func GetMembersZeroMessages(m *messages.Message) {
	members, err := m.Bot.DB.Member.GetByChat(m.Chat.ChatID)
	if err != nil {
		m.Reply("Erro: "+err.Error(), emojis.Fail)
		return
	}

	msg := "Membros com 0 mensagens: \n\n"

	for i, member := range members {
		if member.Messages > 0 {
			continue
		}
		msg += fmt.Sprintf(`- %s`, member.JID)
		if i < len(members)-1 {
			msg += "\n"
		}
	}

	m.Reply(msg, emojis.Success)
}

func MuteMember(m *messages.Message) {

	for _, member := range m.MentionedMembers {
		member.Silenced = 1
		m.Bot.DB.Member.Update(member)
	}

	m.React(emojis.Success)

}

func UnmuteMember(m *messages.Message) {

	for _, member := range m.MentionedMembers {
		member.Silenced = 0
		m.Bot.DB.Member.Update(member)
	}

	m.React(emojis.Success)

}
