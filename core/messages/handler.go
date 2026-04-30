package messages

import (
	"context"

	"github.com/kamuridesu/rainbot-go/internal/bot"
)

type Handler struct {
	Ctx            context.Context
	Bot            *bot.Bot
	CommandHandler func(*Message)
	ChatHandler    func(*Message)
}

func NewHandler(ctx context.Context, commandHandler, chatHandler func(*Message)) *Handler {
	return &Handler{ctx, nil, commandHandler, chatHandler}
}

func (h *Handler) AttachBot(b *bot.Bot) {
	h.Bot = b
}
