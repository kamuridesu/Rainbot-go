package fun

import (
	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
)

func init() {
	commands.NewCommand(
		"slot",
		"Teste sua sorte",
		"diversão",
		nil,
		nil,
		false,
		true,
		true,
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
}
