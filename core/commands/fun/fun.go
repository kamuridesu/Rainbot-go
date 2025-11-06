package fun

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

func ChanceDe(m *messages.Message) {
	text := strings.Join(*m.Args, " ")
	if strings.Contains(text, "virgindade") || strings.Contains(text, "virgem") {
		m.Reply("Nenhuma")
		return
	}

	m.Reply(fmt.Sprintf("A chance %s é de %d%%", text, rand.IntN(100)))
}

func Percent(m *messages.Message) {
	text := strings.Join(*m.Args, " ")
	m.Reply(fmt.Sprintf("Você é %d%% %s", rand.IntN(100), text))
}

func Gado(m *messages.Message) {
	gadoSlice := []string{"ultra extreme gado",
		"Gado-Master",
		"Gado-Rei",
		"Gado",
		"Escravo-ceta",
		"Escravo-ceta Maximo",
		"Gacorno?",
		"Jogador De Forno Livre<3",
		"Mestre Do Frifai<3<3",
		"Gado-Manso",
		"Gado-Conformado",
		"Gado-Incubado",
		"Gado Deus",
		"Mestre dos Gados",
		"Topa tudo por buceta",
		"Gado Comum",
		"Mini Gadinho",
		"Gado Iniciante",
		"Gado Basico",
		"Gado Intermediario",
		"Gado Avançado",
		"Gado Profisional",
		"Gado Mestre",
		"Gado Chifrudo",
		"Corno Conformado",
		"Corno HiperChifrudo",
		"Chifrudo Deus",
		"Mestre dos Chifrudos"}

	choice := gadoSlice[rand.IntN(len(gadoSlice))]
	m.Reply("Você é "+choice, emojis.Success)
}

func Gay(m *messages.Message) {
	gaySlice := []string{
		"hmm... é hetero😔",
		"+/- boiola",
		"tenho minha desconfiança...😑",
		"é né?😏",
		"é ou não?🧐",
		"é gay🙈",
	}

	m.Reply(gaySlice[rand.IntN(len(gaySlice))], emojis.Success)
}
