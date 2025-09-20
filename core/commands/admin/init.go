package admin

import "github.com/kamuridesu/rainbot-go/core/commands"

func init() {

	commands.NewCommand(
		"setup",
		"Configura o bot",
		"admin",
		&[]string{"config"},
		&[]string{"/setup\nprefixo=!"},
		setup,
		commands.IsAdmin,
	)

	commands.NewCommand(
		"warn",
		"Adiciona um warn nos usuários mencionados",
		"admin",
		&[]string{"avisar"},
		&[]string{"/warn @user"},
		WarnUser,
		commands.IsAdmin,
		commands.IsBotAdmin,
	)

	commands.NewCommand(
		"ban",
		"Bane os usuários mencionados do grupo",
		"admin",
		&[]string{"banir"},
		&[]string{"/ban @user"},
		BanUser,
		commands.IsAdmin,
		commands.IsBotAdmin,
	)

}
