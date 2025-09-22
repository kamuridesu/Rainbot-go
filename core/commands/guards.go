package commands

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/messages"
)

func IsGroup(m *messages.Message) error {
	if !m.IsFromGroup() {
		return errors.New("Chat não é grupo")
	}
	return nil
}

func IsAdmin(m *messages.Message) error {

	slog.Info("checking if is group")
	if err := IsGroup(m); err != nil {
		return err
	}

	slog.Info("checking if is admin")

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

func HasMentionedMembers(m *messages.Message) error {

	if len(m.MentionedMembers) < 1 {
		return errors.New("Nenhum usuário mencionado!")
	}

	return nil

}

func HasArgs(expectedArgsNumber int) func(m *messages.Message) error {
	return func(m *messages.Message) error {

		slog.Info("checking if has args")

		if len(*m.Args) < expectedArgsNumber {
			slog.Info("error artgs")
			return fmt.Errorf("Comando esperava %d argumentos!", expectedArgsNumber)
		}

		slog.Info("has args")

		return nil

	}

}
