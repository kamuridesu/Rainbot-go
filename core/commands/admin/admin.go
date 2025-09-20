package admin

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/database/models"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"go.mau.fi/whatsmeow"
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

	slog.Info("Starting ban users")
	for _, user := range m.MentionedMembers {
		user.Warns++
		slog.Info("Updating count for " + user.JID)
		m.Bot.DB.Member.Update(user)
		if user.Warns >= m.Chat.WarnBanThreshold {
			slog.Info("User has reached warn quota, banning")
			toBeBanned = append(toBeBanned, user)
			slog.Info("User scheduled to be banned!")
			continue
		}
		message += fmt.Sprintf("User %s tem %d avisos, mais %d e será banido!\n", user.JID, user.Warns, m.Chat.WarnBanThreshold-user.Warns)
		slog.Info("Successfully warned user")
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
