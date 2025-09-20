package admin

import (
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
)

func setup(m *messages.Message) {

	if len(*m.Args) < 1 {
		config := utils.GetHumanReadableSetup(m.Chat)
		m.Reply(config, emojis.Success)
		return
	}

	err := utils.ParseSetupText(strings.Join(*m.Args, "\n"), m.Chat, m.Bot.DB.Chat)
	if err != nil {
		m.Reply(err.Error(), emojis.Fail)
		return
	}

	m.Reply("Alterações salvas com sucesso", emojis.Success)

}
