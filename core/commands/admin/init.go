package admin

import (
	"github.com/kamuridesu/rainbot-go/core/commands"
)

func init() {

	commands.NewCommand(
		"setup",
		"Configura o bot",
		"admin",
		&[]string{"config"},
		&[]string{"$prefix$alias\nprefixo=!"},
		Setup,
		commands.IsAdmin,
	)

	commands.NewCommand(
		"warn",
		"Adiciona um warn nos usuários mencionados",
		"admin",
		&[]string{"avisar"},
		&[]string{"$prefix$alias @user"},
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
		&[]string{"$prefix$alias @user"},
		RemoveUserWarn,
		commands.IsAdmin,
	)

	commands.NewCommand(
		"ban",
		"Bane os usuários mencionados do grupo",
		"admin",
		&[]string{"banir"},
		&[]string{"$prefix$alias @user"},
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
		&[]string{"$prefix$alias aoba"},
		MentionMembers,
		commands.IsAdmin,
	)

}
