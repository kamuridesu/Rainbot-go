package rucoy

import "github.com/kamuridesu/rainbot-go/core/commands"

var (
	RucoyMenuBannerPath = "assets/rucoy/menu-banner.png"
	RucoyCategory       = commands.NewCategory("rucoy", &RucoyMenuBannerPath)
)

func init() {
	commands.NewCommand("online",
		"Mostra quais jogadores de uma guilda do Rucoy estao online no momento.\n\nUso:\n/online Nome-da-Guild",
		RucoyCategory,
		&[]string{"ronline"},
		&[]string{"${prefix}${alias} Nome-da-Guild", "${prefix}${alias} B L A C K O U T"},
		true, true, false,
		RucoyOnlineGuild,
		commands.HasArgs(1),
	)

	commands.NewCommand("upskill",
		"Calcula quanto tempo falta para subir de uma skill atual ate uma skill desejada usando o tickrate informado.\n\nSe voce informar uma classe (kina, pally ou mage), o bot tambem calcula gasto de mana, potions, gold e envia o card em imagem.\n\nVoce tambem pode informar quantas horas treina por dia.\n\nUso:\n/upskill skill_atual skill_desejada tickrate [classe] [horas_por_dia]\n\nClasses:\nkina = 50 mana por skill\npally = 50 mana por skill + flechas\nmage = 40 mana por skill",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 400 450 42000", "${prefix}${alias} 400 450 42000 kina", "${prefix}${alias} 400 450 42000 pally 8", "${prefix}${alias} 400 450 42000 8 mage"},
		true, true, false,
		Upskill,
		commands.HasArgs(3),
	)

	commands.NewCommand("uplevel",
		"Calcula quanto tempo falta para subir de um level atual ate um level desejado usando sua media de XP por hora.\n\nVoce pode informar quantas horas joga por dia para o bot converter o tempo total em dias de treino.\n\nUso:\n/uplevel level_atual level_desejado xp_por_hora [horas_por_dia]",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 350 400 20kk", "${prefix}${alias} 350 400 20kk 8", "${prefix}${alias} 275 300 5kk"},
		true, true, false,
		Uplevel,
		commands.HasArgs(3),
	)

	commands.NewCommand("train",
		"Calcula o melhor monstro para AFK Train e Power Train baseado na arma, level, skill e add informado.\n\nO add pode ser negativo, por exemplo -50, para simular item que reduz skill.\n\nVoce tambem pode informar a eficiencia minima desejada.\n\nUso:\n/train arma level skill add [eficiencia]\n\nArmas de treino comuns:\n4, 5, 7, 9, 11 e 13",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 5 351 391 -50", "${prefix}${alias} 5 351 391 -50 90", "${prefix}${alias} 7 400 450 0"},
		true, true, false,
		RucoyTrain,
		commands.HasArgs(4),
	)

	commands.NewCommand("afk",
		"Mostra jogadores de uma guilda que estao ha 7 dias ou mais sem logar no Rucoy.\n\nUso:\n/afk Nome-da-Guild",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} Nome-da-Guild", "${prefix}${alias} B L A C K O U T"},
		true, true, false,
		RucoyAFKGuild,
		commands.HasArgs(1),
	)

	commands.NewCommand("info",
		"Mostra informacoes de um jogador do Rucoy: nome, level, guild, titulo, status online ou ultima vez online, Black Skull e XP Mobwin.\n\nUso:\n/info Nome-do-Jogador",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} Nome-do-Jogador", "${prefix}${alias} Kamuri SG"},
		true, true, false,
		RucoyInfo,
		commands.HasArgs(1),
	)

	commands.NewCommand("meta",
		"Mostra quais membros de uma guilda ainda nao chegaram em uma meta de level.\n\nO bot lista apenas quem ainda nao bateu a meta e quantos levels faltam.\n\nUso:\n/meta level_meta Nome-da-Guild",
		RucoyCategory,
		nil,
		&[]string{"${prefix}${alias} 400 Nome-da-Guild", "${prefix}${alias} 500 B L A C K O U T"},
		true, true, false,
		RucoyMetaGuild,
		commands.HasArgs(2),
	)
}
