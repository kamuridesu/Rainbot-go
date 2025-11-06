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
		return err
	}

	gInfo, err := m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		return errors.New("Algum erro ocorreu: " + err.Error())
	}

	for _, participant := range gInfo.Participants {
		if (participant.LID.ToNonAD().String() == m.Author.JID) && participant.IsAdmin {
			return nil
		}
	}

	return errors.New("Comando pode ser usado apenas por admins")
}

func IsBotAdmin(m *messages.Message) error {

	botJid := m.Bot.Client.Store.GetLID().ToNonAD().String()

	gInfo, err := m.Bot.Client.GetGroupInfo(m.RawEvent.Info.Chat)
	if err != nil {
		return errors.New("Algum erro ocorreu: " + err.Error())
	}

	for _, participant := range gInfo.Participants {
		if participant.LID.ToNonAD().String() == botJid && participant.IsAdmin {
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

// Checks for __at least__ n number of arguments
// Returns a function thats the traditional guard as expected by the runner.
func HasArgs(atLeast int) func(m *messages.Message) error {
	return func(m *messages.Message) error {
		if atLeast == 0 {
			return nil
		}

		if len(*m.Args) < atLeast {
			return fmt.Errorf("Comando esperava %d argumentos!", atLeast)
		}

		return nil

	}

}

// Checks for quoted message
func HasQuotedMessage(m *messages.Message) error {
	if m.QuotedMessage == nil {
		return errors.New("Este comando só pode ser usado ao mencionar uma mensagem.")
	}
	return nil
}
