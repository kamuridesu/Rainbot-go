package v1

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/kamuridesu/gomechan/core/routes"
)

func Serve(addr string) {
	mux := http.NewServeMux()
	routes.AddHealthCheck(mux)

	slog.Info("Listening on " + addr)
	err := http.ListenAndServe(addr, mux)
	if errors.Is(err, http.ErrServerClosed) {
		slog.Error("server closed")
	} else if err != nil {
		slog.Error("unkown error", "error", err)
		os.Exit(1)
	}
}
