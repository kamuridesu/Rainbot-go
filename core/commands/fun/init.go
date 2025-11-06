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
		NewSticker,
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

}
