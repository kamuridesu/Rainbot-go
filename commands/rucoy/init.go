package rucoy

import "github.com/kamuridesu/rainbot-go/core/commands"

var (
	RucoyCategory = commands.NewCategory("rucoy", nil)
)

func init() {
	commands.NewCommand("online",
		"Mostra os membros online de uma guilda no Rucoy online",
		RucoyCategory,
		&[]string{"ronline"},
		nil,
		true, true, false,
		RucoyOnlineGuild,
		commands.HasArgs(1),
	)

	commands.NewCommand("upskill",
		"Calcula quanto tempo demora para ir de uma skill ate outra no Rucoy Online",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 400 450 5000", "${prefix}${alias} 400 450 5000 kina 8"},
		true, true, false,
		Upskill,
		commands.HasArgs(3),
	)

	commands.NewCommand("uplevel",
		"Calcula quanto tempo demora para ir de um level ate outro no Rucoy Online",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 350 400 20kk"},
		true, true, false,
		Uplevel,
		commands.HasArgs(3),
	)

	commands.NewCommand("train",
		"Calcula o melhor monstro para AFK train e powertrain no Rucoy Online",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 5 351 391 -50"},
		true, true, false,
		RucoyTrain,
		commands.HasArgs(4),
	)

	commands.NewCommand("afk",
		"Mostra jogadores de uma guilda do Rucoy com 7 dias ou mais sem logar",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} B L A C K O U T"},
		true, true, false,
		RucoyAFKGuild,
		commands.HasArgs(1),
	)

	commands.NewCommand("info",
		"Mostra informacoes de um jogador do Rucoy Online",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} Nome do Jogador"},
		true, true, false,
		RucoyInfo,
		commands.HasArgs(1),
	)

	commands.NewCommand("meta",
		"Mostra membros de uma guilda do Rucoy que ainda nao bateram uma meta de level",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 400 B L A C K O U T"},
		true, true, false,
		RucoyMetaGuild,
		commands.HasArgs(2),
	)
}
