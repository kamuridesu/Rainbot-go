package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	b "github.com/kamuridesu/rainbot-go/internal/bot"
)

func main() {
	ctx := context.Background()
	commands.RegisterHelpMenu()
	handler := messages.NewHandler(ctx, commands.RunCommand)
	bot, err := b.New(ctx, "Teto", "!", "sqlite3", "file:examplestore.db?_foreign_keys=on", handler)
	if err != nil {
		panic(err)
	}
	defer bot.Disconnect()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

}
