package offenses

import (
	"math/rand/v2"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
)

var responses = []string{
	"Sua mãe",
	"Vsf otario",
	"Mama aqui arrombado",
	"Ah vai dar meia hora de cu",
	"E dai",
	"Perguntei?",
	"Sei q tu quer me dar",
	"Quando eu te comi tu tava manso ne",
	"Vc",
	"Seu rabo",
}

func OffendsBot(m *messages.Message) bool {

	if m.Chat.ProfanityFilterEnabled == 1 {
		return false
	}

	text := strings.ToLower(*m.Text)
	if !strings.Contains(text, "bot") {
		return false
	}
	contains := false
	for _, value := range []string{"inutil", "inútil", "lixo", "arrombado", "fdp", "ruim"} {
		if strings.Contains(text, value) {
			contains = true
			break
		}
	}
	if contains {
		choice := responses[rand.IntN(len(responses))]
		m.Reply(choice)
		return true
	}
	return false
}
