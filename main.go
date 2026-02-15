package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	v1 "github.com/kamuridesu/rainbot-go/core/api/v1"
	"github.com/kamuridesu/rainbot-go/core/chat"
	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
	b "github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/database"
	"github.com/kamuridesu/rainbot-go/internal/utils"

	_ "github.com/kamuridesu/rainbot-go/core/commands/admin"
	_ "github.com/kamuridesu/rainbot-go/core/commands/fun"
)

func init() {
	utils.ReadDotEnv()

	slog.Info("Starting migrations...")
	utils.Migrate()
	slog.Info("All migrations have finished!")
}

func main() {

	defer func() {
		if p := recover(); p != nil {
			slog.Error(fmt.Sprintf("Fatal error: %v", p))
			os.Exit(1)
		}
	}()

	singleton, err := database.InitDatabaseSingleton(os.Getenv("DB_DRIVER"), os.Getenv("DB_PARAMS"))
	if err != nil {
		panic(err)
	}
	defer singleton.Close()

	ctx := context.Background()

	slog.Info(fmt.Sprintf("Found %d commands", len(*commands.GetLoadedCommands())))

	handler := messages.NewHandler(ctx, commands.RunCommand, chat.ChatHandler)
	botName := os.Getenv("BOT_NAME")
	if botName == "" {
		botName = "teto"
	}

	bot, err := b.New(ctx, botName, os.Getenv("DB_DRIVER"), os.Getenv("DB_PARAMS"), handler, singleton)
	if err != nil {
		panic(err)
	}
	defer bot.Disconnect()

	go func() {
		v1.Serve(":8080")
	}()

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

}
