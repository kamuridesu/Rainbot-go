package test

import (
	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
)

func init() {
	commands.NewCommand("test", "", "test", nil, nil, false, false, false, func(m *messages.Message) {
		m.Bot.Disconnect()
	})
}
