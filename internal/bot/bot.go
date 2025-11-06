package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kamuridesu/rainbot-go/internal/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	Name          *string
	Client        *whatsmeow.Client
	DB            *database.DatabaseSingleton
	StartTime     time.Time
	CreatorNumber *string
}

type Handler interface {
	Handle(any)
	AttachBot(*Bot)
}

func New(ctx context.Context, name, sqlDialact, dbAddress string, handler Handler, singleton *database.DatabaseSingleton) (*Bot, error) {

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
	client.ManualHistorySyncDownload = true

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
	tmpc := os.Getenv("CREATOR_NUMBER")
	creatorN := &tmpc
	if tmpc == "" {
		creatorN = nil
	}

	bot := Bot{
		Name:          &name,
		Client:        client,
		DB:            singleton,
		StartTime:     time.Now(),
		CreatorNumber: creatorN,
	}
	client.AddEventHandler(handler.Handle)
	handler.AttachBot(&bot)
	slog.Info("New client connected")

	return &bot, nil
}

func (b *Bot) Disconnect() {
	b.Client.Disconnect()
}
