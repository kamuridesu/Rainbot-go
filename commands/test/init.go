package test

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/chatter"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

var (
	TestCategory = commands.NewCategory("test", nil)
)

func init() {
	commands.NewCommand("test", "", TestCategory, nil, nil, false, false, false, func(m *messages.Message) {
		fmt.Println("rcvd test command")
		c := chatter.New(m.Bot.DB.AIModel)
		fmt.Println("chatter instantiated")
		msg, err := c.GenerateMessage(m.Ctx, *m.Bot.Name, m.Chat.Prefix, strings.Join(*m.Args, " "))
		if err != nil {
			slog.Error(err.Error())
			m.Reply("error generating response: "+err.Error(), emojis.Fail)
		}
		m.Reply(msg)
	})
}
