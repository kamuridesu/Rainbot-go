package profanity

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/kamuridesu/rainbot-go/internal/database/models"
)

func CheckCustomWord(chat *models.Chat, text string) error {

	if chat.CustomProfanityWords == "" {
		return nil
	}

	wordlist := strings.Split(chat.CustomProfanityWords, ",")
	if len(wordlist) == 0 {
		return nil
	}
	for term := range strings.SplitSeq(strings.ToLower(text), " ") {
		if slices.Contains(wordlist, term) {
			slog.Warn(fmt.Sprintf("CCW: User usou palavra proibida: %s", term))
			return errors.New("Ei! Você falou uma palavra proibida, cadê sua educação?")
		}
	}
	return nil

}
