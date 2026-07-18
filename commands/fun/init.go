package fun

import (
	"strings"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/lyrics"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

func init() {
	commands.NewCommand(
		"slot",
		"Teste sua sorte",
		"diversão",
		nil,
		nil,
		true,
		true,
		false,
		Slot,
	)

	commands.NewCommand(
		"filter",
		"Adiciona filtros ou visualiza, um filtro é uma resposta automatica a uma mensagem.",
		"diversão",
		&[]string{"filtro"},
		&[]string{"${prefix}${alias} antedeguemon"},
		true,
		false,
		false,
		func(m *messages.Message) {
			if len(*m.Args) < 1 {
				ShowFilters(m)
				return
			}
			NewFilter(m)
		},
	)

	commands.NewCommand(
		"delfilter",
		"Deleta um filter existente",
		"diversão",
		&[]string{"rfilter", "rmfilter", "deletafilter"},
		nil,
		true,
		false,
		false,
		DeleteFilter,
		commands.HasArgs(1),
	)

	commands.NewCommand(
		"sticker",
		"Cria um sticker",
		"diversão",
		&[]string{"fig", "s", "figu"},
		nil,
		true,
		false,
		false,
		NewStickerSquash,
	)

	commands.NewCommand(
		"ssticker",
		"Cria um sticker com o tamanho original",
		"diversão",
		&[]string{"figs", "ss", "figus"},
		nil,
		true,
		false,
		false,
		NewStickerOriginal,
	)

	commands.NewCommand(
		"casal",
		"Sorteia 2 membros e forma um casal",
		"diversão",
		nil, nil, true, false, false,
		Casal,
		commands.IsGroup,
	)

	commands.NewCommand(
		"copy",
		"Copia uma mensagem em visualização única e envia no privado",
		"diversão",
		&[]string{"c", "copiar", "revelar", "reveal"},
		nil,
		true, false, false,
		RevealMessage,
		commands.HasQuotedMessage,
	)

	commands.NewCommand("chance",
		"Calcula a chance de algo",
		"diversão",
		nil,
		&[]string{"${prefix}chance de eu ficar rico"},
		true, false, false, ChanceDe, commands.HasArgs(1))

	commands.NewCommand("percent",
		"Calcula o quanto % você é",
		"diversão",
		&[]string{"perc"},
		&[]string{"${prefix}${alias} pobre"},
		true, false, false, Percent, commands.HasArgs(1))

	commands.NewCommand("gado", "Diz o seu nivel de gado", "diversão", nil, nil, true, false, false, Gado)
	commands.NewCommand("gay", "Diz o seu nivel de gay", "diversão", nil, nil, true, false, false, Gay)

	commands.NewCommand("lyrics",
		"Envia a letra de alguma música",
		"diversão",
		&[]string{"letras"},
		&[]string{"${prefix}${alias} never gonna give you up"},
		true, false, false,
		func(m *messages.Message) {
			args := strings.Join(*m.Args, " ")
			l, err := lyrics.SearchLyrics(m.Ctx, args)
			if err != nil {
				m.Reply("Falha ao pesquisar a musica: "+err.Error(), emojis.Fail)
				return
			}
			m.Reply(*l, emojis.Success)
		},
	)

	commands.NewCommand("video", "Baixa um video", "diversão", nil, nil, true, false, false, DownloadVideo, commands.HasArgs(1, true))
	commands.NewCommand("music", "Baixa uma musica", "diversão", nil, nil, true, false, false, DownloadAudio, commands.HasArgs(1, true))

	commands.NewHiddenCommand("rucoy",
		"Mostra o menu de comandos do Rucoy Online",
		"diversÃ£o",
		nil,
		nil,
		true, true, false,
		RucoyMenu,
	)

	commands.NewCommand("ronline",
		"Mostra os membros online de uma guilda no Rucoy online",
		"diversão",
		&[]string{"online"},
		nil,
		true, true, false,
		RucoyOnlineGuild,
		commands.HasArgs(1),
	)

	commands.NewCommand("upskill",
		"Calcula quanto tempo demora para ir de uma skill ate outra no Rucoy Online",
		"diversão",
		nil,
		&[]string{"${prefix}${alias} 400 450 5000"},
		true, true, false,
		Upskill,
		commands.HasArgs(3),
	)

	commands.NewCommand("uplevel",
		"Calcula quanto tempo demora para ir de um level ate outro no Rucoy Online",
		"diversÃ£o",
		nil,
		&[]string{"${prefix}${alias} 350 400 20kk"},
		true, true, false,
		Uplevel,
		commands.HasArgs(3),
	)

	commands.NewCommand("train",
		"Calcula o melhor monstro para AFK train e powertrain no Rucoy Online",
		"diversÃ£o",
		nil,
		&[]string{"${prefix}${alias} 5 351 391 -50"},
		true, true, false,
		RucoyTrain,
		commands.HasArgs(4),
	)

	commands.NewCommand("afk",
		"Mostra jogadores de uma guilda do Rucoy com 7 dias ou mais sem logar",
		"diversão",
		nil,
		&[]string{"${prefix}${alias} B L A C K O U T"},
		true, true, false,
		RucoyAFKGuild,
		commands.HasArgs(1),
	)

	commands.NewCommand("meta",
		"Mostra membros de uma guilda do Rucoy que ainda nao bateram uma meta de level",
		"diversÃ£o",
		nil,
		&[]string{"${prefix}${alias} 400 B L A C K O U T"},
		true, true, false,
		RucoyMetaGuild,
		commands.HasArgs(2),
	)

	commands.NewCommand("quotly",
		"gera um quote",
		"diversão",
		&[]string{"q"},
		nil,
		true, false, false,
		HandleQuoteCommand,
		commands.HasQuotedMessage,
	)

	commands.NewCommand("randomquotly",
		"envia um quote aleatorio",
		"diversão",
		&[]string{"qrand"},
		nil,
		true, false, false,
		RandomQuote,
	)
}
