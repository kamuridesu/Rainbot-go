package chat

import (
	"log/slog"

	"github.com/kamuridesu/rainbot-go/core/chat/filter"
	"github.com/kamuridesu/rainbot-go/core/messages"
)

func ChatHandler(m *messages.Message) {
	slog.Info("Handling non command msg")
	filter.GetChatFilters(m)
}
