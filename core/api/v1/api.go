package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/kamuridesu/gomechan/core/routes"
	"github.com/kamuridesu/rainbot-go/internal/bot"
)

func Serve(addr string, b *bot.Bot) {
	mux := http.NewServeMux()
	routes.AddHealthCheck(mux, b.IsAlive)

	slog.Info("Listening on " + addr)
	err := http.ListenAndServe(addr, mux)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Error("server closed")
	} else if err != nil {
		slog.Error("unkown error", "error", err)
		os.Exit(1)
	}
}
