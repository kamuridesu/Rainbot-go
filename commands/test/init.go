package test

import (
	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
)

var (
	TestCategory = commands.NewCategory("test", nil)
)

func init() {
	commands.NewCommand("test", "", TestCategory, nil, nil, false, false, false, func(m *messages.Message) {
		m.Bot.Disconnect()
	})
}
