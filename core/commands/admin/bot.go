package admin

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func Setup(m *messages.Message) {

	if len(*m.Args) < 1 {
		config := utils.GetHumanReadableSetup(m.Chat)
		m.Reply(config, emojis.Success)
		return
	}

	err := utils.ParseSetupText(*m.Args, m.Chat, m.Bot.DB.Chat)
	if err != nil {
		m.Reply(err.Error(), emojis.Fail)
		return
	}

	m.Reply("Alterações salvas com sucesso", emojis.Success)

}

func Bug(m *messages.Message) {
	if m.Bot.CreatorNumber == nil {
		m.Reply("Nenhum numero configurado.", emojis.Fail)
	}
	message := fmt.Sprintf("Bug reportado: \n\n%s", strings.Join(*m.Args, " "))

	_, err := m.SendMessage(&waE2E.Message{
		Conversation: &message,
	}, types.NewJID(*m.Bot.CreatorNumber, types.DefaultUserServer))

	if err != nil {
		slog.Error(err.Error())
		m.Reply("falha ao enviar: "+err.Error(), emojis.Fail)
		return
	}
	m.React(emojis.Success)

}
