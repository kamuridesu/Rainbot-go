package admin

import (
	"fmt"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

func init() {

	commands.NewCommand(
		"setup",
		"Configura o bot",
		"admin",
		&[]string{"config"},
		&[]string{"${prefix}${alias}\nprefixo=!"},
		false,
		false,
		false,
		Setup,
		commands.IsAdmin,
	)

	commands.NewCommand(
		"warn",
		"Adiciona um warn nos usuários mencionados",
		"admin",
		&[]string{"avisar"},
		&[]string{"${prefix}${alias} @user"},
		false,
		false,
		false,
		WarnUser,
		commands.IsAdmin,
		commands.IsBotAdmin,
		commands.HasMentionedMembers,
	)

	commands.NewCommand(
		"removewarn",
		"Remove um aviso de um membro",
		"admin",
		&[]string{"rwarn"},
		&[]string{"${prefix}${alias} @user"},
		false,
		false,
		false,
		RemoveUserWarn,
		commands.IsAdmin,
		commands.HasMentionedMembers,
	)

	commands.NewCommand(
		"ban",
		"Bane os usuários mencionados do grupo",
		"admin",
		&[]string{"banir"},
		&[]string{"${prefix}${alias} @user"},
		false,
		false,
		false,
		BanUser,
		commands.IsAdmin,
		commands.IsBotAdmin,
		commands.HasMentionedMembers,
	)

	commands.NewCommand(
		"todos",
		"Menciona os membros do grupo",
		"admin",
		&[]string{"all"},
		&[]string{"${prefix}${alias} aoba"},
		false,
		false,
		false,
		MentionMembers,
		commands.IsAdmin,
	)

	commands.NewCommand(
		"promover",
		"Da permissão de admin dos usuários mencionados",
		"admin",
		&[]string{"promote"},
		&[]string{"${prefix}${alias} @user"},
		false,
		false,
		false,
		func(m *messages.Message) {
			err := changeUserAdminStatus(m)
			if err != nil {
				m.Reply(fmt.Sprintf("Erro: %s", err), emojis.Fail)
				return
			}
			m.Reply("Usuário(s) promovidos com sucesso", emojis.Success)
		},
		commands.IsAdmin,
		commands.IsBotAdmin,
		commands.HasMentionedMembers,
	)

	commands.NewCommand(
		"rebaixar",
		"Remove o admin dos usuários mencionados",
		"admin",
		&[]string{"demote"},
		&[]string{"${prefix}${alias} @user"},
		false,
		false,
		false,
		func(m *messages.Message) {
			err := changeUserAdminStatus(m, true)
			if err != nil {
				m.Reply(fmt.Sprintf("Erro: %s", err), emojis.Fail)
				return
			}
			m.Reply("Usuário(s) rebaixados com sucesso", emojis.Success)
		},
		commands.IsAdmin,
		commands.IsBotAdmin,
		commands.HasMentionedMembers,
	)

	commands.NewCommand("msg",
		"Lista as mensagens enviadas por membros do grupo, ou membros sem mensagens", "admin",
		&[]string{"lmsg", "mensagens"},
		&[]string{
			"${prefix}${alias} zero mostra membros com 0 mensgens",
			"${prefix}${alias} limpar limpa as mensagens do grupo",
			"${prefix}${alias} mostra as mensagens por membro"}, false, false, false, func(m *messages.Message) {
			if m.Args != nil && len(*m.Args) > 0 {
				text := strings.Join(*m.Args, " ")
				switch text {
				case "limpar":
					if err := commands.IsAdmin(m); err != nil {
						m.Reply(err.Error(), emojis.Fail)
						return
					}
					PurgeMessages(m)
					return
				case "zero":
					if err := commands.IsAdmin(m); err != nil {
						m.Reply(err.Error(), emojis.Fail)
						return
					}
					GetMembersZeroMessages(m)
					return
				}
			}
			MessagesPerMember(m)
		}, commands.IsGroup)

	commands.NewCommand("mute",
		"Silencia membros mencionados", "admin", nil, &[]string{"${prefix}${alias} @user"},
		false, false, false, MuteMember,
		commands.HasMentionedMembers,
		commands.IsGroup,
		commands.IsAdmin,
		commands.IsBotAdmin,
	)

	commands.NewCommand("unmute", "Deixa os membros falarem novamente", "admin", &[]string{"um"}, &[]string{"${prefix}${alias} @user"},
		false, false, false,
		UnmuteMember,
		commands.HasMentionedMembers, commands.IsGroup, commands.IsAdmin)

	commands.NewCommand("bug", "Reporta um bug", "misc", nil, nil, false, false, false, Bug, commands.HasArgs(1))
	commands.NewCommand("transmitir",
		"Transmite uma mensagem",
		"misc", &[]string{"bc", "broadcast"}, nil, false, false, false, Broadcast, commands.HasArgs(2))

}
