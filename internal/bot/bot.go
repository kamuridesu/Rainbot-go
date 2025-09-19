package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	Name   *string
	Prefix *string
	Client *whatsmeow.Client
}

type Handler interface {
	Handle(any)
	AttachBot(*Bot)
}

func New(ctx context.Context, name, prefix, sqlDialact, dbAddress string, handler Handler) (*Bot, error) {

	dblog := waLog.Stdout("Database", "INFO", true)
	container, err := sqlstore.New(ctx, sqlDialact, dbAddress, dblog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, err
	}

	clientLog := waLog.Stdout("Client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	slog.Info("Starting new client")
	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QRCode: ")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event: ", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			return nil, err
		}
	}
	bot := Bot{
		Name:   &name,
		Prefix: &prefix,
		Client: client,
	}
	client.AddEventHandler(handler.Handle)
	handler.AttachBot(&bot)
	slog.Info("New client connected")

	return &bot, nil
}

func (b *Bot) Disconnect() {
	b.Client.Disconnect()
}
