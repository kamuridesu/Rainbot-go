package bot_test

import (
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
)

func TestBotIsAliveWhenConnected(t *testing.T) {
	fake := &botfakes.FakeClient{IsConnectedFunc: func() bool { return true }}
	b := &bot.Bot{Client: fake}

	if err := b.IsAlive(); err != nil {
		t.Errorf("IsAlive() error = %v, want nil", err)
	}
}

func TestBotIsAliveWhenDisconnected(t *testing.T) {
	fake := &botfakes.FakeClient{IsConnectedFunc: func() bool { return false }}
	b := &bot.Bot{Client: fake}

	if err := b.IsAlive(); err == nil {
		t.Error("expected an error when the client reports disconnected")
	}
}

func TestBotDisconnect(t *testing.T) {
	fake := &botfakes.FakeClient{}
	b := &bot.Bot{Client: fake}

	b.Disconnect()

	if !fake.DisconnectCalled {
		t.Error("expected Disconnect() to be called on the client")
	}
}
