package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/kamuridesu/rainbot-go/core/database"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Bot struct {
	Name          *string
	Client        WhatsAppClient
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
		Client:        &whatsmeowClientAdapter{client},
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

func (b *Bot) IsAlive() error {
	if !b.Client.IsConnected() {
		return errors.New("wa client is offline")
	}
	return nil
}

func (b *Bot) IsSelf(jid types.JID) bool {
	device := b.Client.Store()
	if device == nil {
		return false
	}
	jid = jid.ToNonAD()
	if device.ID != nil && jid == device.ID.ToNonAD() {
		return true
	}
	return !device.LID.IsEmpty() && jid == device.LID.ToNonAD()
}

func (b *Bot) ContainsSelf(jids []types.JID) bool {
	for _, jid := range jids {
		if b.IsSelf(jid) {
			return true
		}
	}
	return false
}

// ToLID returns the LID form of jid, which is how members are stored, or an
// empty string if a phone number JID has no known LID.
func (b *Bot) ToLID(ctx context.Context, jid types.JID) string {
	jid = jid.ToNonAD()
	if jid.Server == types.HiddenUserServer {
		return jid.String()
	}
	device := b.Client.Store()
	if device == nil || device.LIDs == nil {
		return ""
	}
	lid, err := device.LIDs.GetLIDForPN(ctx, jid)
	if err != nil || lid.IsEmpty() {
		return ""
	}
	return lid.ToNonAD().String()
}
