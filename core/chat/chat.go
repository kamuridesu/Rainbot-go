package chat

import (
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/chat/filter"
	"github.com/kamuridesu/rainbot-go/core/chat/mute"
	"github.com/kamuridesu/rainbot-go/core/chat/offenses"
	"github.com/kamuridesu/rainbot-go/core/chat/profanity"
	"github.com/kamuridesu/rainbot-go/core/messages"
)

func ChatHandler(m *messages.Message) {
	if mute.DeleteIfMuted(m) {
		slog.Info("Deleted muted message")
		return
	}
	if profanity.CheckForWord(m) {
		slog.Info("Blocked word caught")
		return
	}
	if offenses.OffendsBot(m) {
		slog.Info("Someone offended the bot")
		return
	}
	if err := filter.GetChatFilters(m); err != nil {
		slog.Error(err.Error())
		return
	}
}
