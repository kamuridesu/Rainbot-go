package fun

import (
	"fmt"
	"math/rand/v2"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

const (
	FRUITS = "🥑🍉🍓🍎🍍🥝🍑🥥🍋🍐🍌🍒🔔🍊🍇"
)

func Slot(m *messages.Message) {
	fruits := []rune(FRUITS)
	numFruits := len(fruits)

	var selected []string
	fruitText := ""
	for i := range 3 {
		randomIndex := rand.IntN(numFruits)
		selectedFruit := string(fruits[randomIndex])
		selected = append(selected, selectedFruit)
		fruitText += selectedFruit
		if i != 2 {
			fruitText += " : "
		}
	}

	didWin := selected[0] == selected[1] && selected[1] == selected[2] && selected[2] == selected[0]

	rawMsg := `Consiga 3 iguais para ganhar
╔═══ ≪ •❈• ≫ ════╗
║  [💰SLOT💰 | 777 ]
║
║
║      %s  ◄━━┛
║
║
║  [💰SLOT💰 | 777 ]
╚════ ≪ •❈• ≫ ═══╝

%s`

	if didWin {
		message := fmt.Sprintf(rawMsg, fruitText, "Parabéns! Você ganhou!")
		m.Reply(message, emojis.Success)
	} else {
		message := fmt.Sprintf(rawMsg, fruitText, "Que pena! Tente novamente.")
		m.Reply(message, emojis.Fail)
	}

}
