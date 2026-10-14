package bot_test

import (
	"context"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/types"
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

func TestBotIsSelfMatchesPhoneAndLIDIgnoringDevice(t *testing.T) {
	id := types.NewADJID("5511999999999", 0, 12)
	device := &store.Device{ID: &id, LID: types.NewADJID("123456", 0, 12)}
	device.LID.Server = types.HiddenUserServer
	b := &bot.Bot{Client: &botfakes.FakeClient{StoreFunc: func() *store.Device { return device }}}

	cases := map[types.JID]bool{
		types.NewJID("5511999999999", types.DefaultUserServer): true,
		types.NewJID("123456", types.HiddenUserServer):         true,
		types.NewJID("5511888888888", types.DefaultUserServer): false,
		types.NewJID("123456", types.DefaultUserServer):        false,
	}
	for jid, want := range cases {
		if got := b.IsSelf(jid); got != want {
			t.Errorf("IsSelf(%s) = %v, want %v", jid, got, want)
		}
	}

	if b.ContainsSelf([]types.JID{types.NewJID("5511888888888", types.DefaultUserServer)}) {
		t.Error("ContainsSelf() = true for a list without the bot")
	}
}

func TestBotIsSelfWithoutStore(t *testing.T) {
	b := &bot.Bot{Client: &botfakes.FakeClient{}}
	if b.IsSelf(types.NewJID("5511999999999", types.DefaultUserServer)) {
		t.Error("IsSelf() = true with no device store")
	}
}

func TestBotToLIDKeepsLIDsAndDropsDevice(t *testing.T) {
	b := &bot.Bot{Client: &botfakes.FakeClient{}}
	jid := types.NewADJID("123456", 0, 12)
	jid.Server = types.HiddenUserServer

	if got := b.ToLID(context.Background(), jid); got != "123456@lid" {
		t.Errorf("ToLID() = %q, want %q", got, "123456@lid")
	}
}

func TestBotToLIDWithoutStoreReturnsEmpty(t *testing.T) {
	b := &bot.Bot{Client: &botfakes.FakeClient{}}
	if got := b.ToLID(context.Background(), types.NewJID("5511999999999", types.DefaultUserServer)); got != "" {
		t.Errorf("ToLID() = %q, want empty", got)
	}
}
