package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	b "github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/database"

	_ "github.com/kamuridesu/rainbot-go/core/commands/admin"
)

func main() {
	singleton, err := database.InitDatabaseSingleton("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	defer singleton.Close()

	ctx := context.Background()
	commands.RegisterHelpMenu()

	slog.Info(fmt.Sprintf("FOund %d commands", len(*commands.GetLoadedCommands())))

	for _, command := range *commands.GetLoadedCommands() {
		slog.Info("Command: " + command.Name)
	}

	handler := messages.NewHandler(ctx, commands.RunCommand)

	bot, err := b.New(ctx, "Teto", "sqlite3", "file:examplestore.db?_foreign_keys=on", handler, singleton)
	if err != nil {
		panic(err)
	}
	defer bot.Disconnect()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

}
