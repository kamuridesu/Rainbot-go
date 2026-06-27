package quotly

import (
	"database/sql"
	"errors"
	"log/slog"
	"math/rand/v2"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/storage"
)

func RandomQuoteDrop(m *messages.Message) {
	if m.Chat.AllowQuote == 0 {
		return
	}

	odds := m.Chat.QuoteNMessages
	if odds <= 0 {
		odds = 300
	}

	if r := rand.IntN(odds); r != 0 {
		return
	}

	randomQuote, err := m.Bot.DB.Quotly.GetRandomByChat(m.Chat.ChatID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Debug("no quotes are saved for this chat yet")
			return
		}
		slog.Error("Failed to fetch random quote from DB", "err", err)
		return
	}

	file := storage.NewFile(randomQuote.FileId, storage.ModeReadOnly)
	exists, err := file.Exists(m.Ctx)
	if err != nil || !exists {
		slog.Warn("file is missing from storage", "fileId", randomQuote.FileId)
		return
	}

	data, err := file.Read(m.Ctx)
	if err != nil {
		slog.Error("Failed to read random quote file from storage", "err", err, "fileId", randomQuote.FileId)
		return
	}

	slog.Info("Sending random quote", "fileId", randomQuote.FileId)
	_, err = m.ReplySticker(data, messages.StickerMessage)
	if err != nil {
		slog.Error("Failed to send random quote sticker", "err", err)
	}
}
