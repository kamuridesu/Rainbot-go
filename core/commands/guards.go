package commands

import (
	"errors"
	"fmt"

	"github.com/kamuridesu/rainbot-go/core/messages"
)

func IsGroup(m *messages.Message) error {
	if !m.IsFromGroup() {
		return errors.New("Chat não é grupo")
	}
	return nil
}

func IsAdmin(m *messages.Message) error {
	if err := IsGroup(m); err != nil {
		return nil
	}

	gInfo, err := m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		return errors.New("Algum erro ocorreu: " + err.Error())
	}

	for _, participant := range gInfo.Participants {
		if (participant.JID.String() == m.Author.JID) && participant.IsAdmin {
			return nil
		}
	}

	return errors.New("Comando pode ser usado apenas por admins")
}

func IsBotAdmin(m *messages.Message) error {

	botJid := m.Bot.Client.Store.GetJID().ToNonAD().String()

	gInfo, err := m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		return errors.New("Algum erro ocorreu: " + err.Error())
	}

	fmt.Printf("Bot id: %v\n", botJid)

	for _, participant := range gInfo.Participants {
		if participant.JID.String() == botJid && participant.IsAdmin {
			return nil
		}
	}

	return errors.New("Bot não é admin")

}
