package admin

import (
	"fmt"
	"log/slog"
	"os"
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

func Broadcast(m *messages.Message) {

	bcPasswd := os.Getenv("BROADCAST_PASSWORD")
	if bcPasswd == "" {
		m.Reply("Nenhuma senha configurada.", emojis.Fail)
		return
	}

	passwd := (*m.Args)[0]
	message := "Transmissão: \n\n" + strings.Join((*m.Args)[1:], " ")

	if passwd != bcPasswd {
		m.Reply("Senha invalida.", emojis.Fail)
		return
	}

	groups, err := m.Bot.Client.GetJoinedGroups(m.Ctx)
	if err != nil {
		m.Reply("Houve um erro ao ler os grupos: "+err.Error(), emojis.Fail)
		return
	}

	for _, group := range groups {
		msg := &waE2E.Message{
			Conversation: &message,
		}
		_, err := m.SendMessage(msg, group.JID)
		if err != nil {
			slog.Error(err.Error())
			continue
		}
	}

}
