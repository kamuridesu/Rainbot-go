package utils

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func TestDownloadIUserProfilePicInvalidJID(t *testing.T) {
	fake := &botfakes.FakeClient{}
	b := &bot.Bot{Client: fake}

	if _, err := DownloadIUserProfilePic(context.Background(), "111:notanumber@lid", b); err == nil {
		t.Error("expected an error for an unparseable JID")
	}
}

func TestDownloadIUserProfilePicGetInfoError(t *testing.T) {
	wantErr := errors.New("no profile pic")
	fake := &botfakes.FakeClient{
		GetProfilePictureInfoFunc: func(ctx context.Context, jid types.JID, params *whatsmeow.GetProfilePictureParams) (*types.ProfilePictureInfo, error) {
			return nil, wantErr
		},
	}
	b := &bot.Bot{Client: fake}

	if _, err := DownloadIUserProfilePic(context.Background(), "111@lid", b); !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestDownloadIUserProfilePicSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-profile-pic-bytes"))
	}))
	defer server.Close()

	fake := &botfakes.FakeClient{
		GetProfilePictureInfoFunc: func(ctx context.Context, jid types.JID, params *whatsmeow.GetProfilePictureParams) (*types.ProfilePictureInfo, error) {
			return &types.ProfilePictureInfo{URL: server.URL}, nil
		},
	}
	b := &bot.Bot{Client: fake}

	got, err := DownloadIUserProfilePic(context.Background(), "111@lid", b)
	if err != nil {
		t.Fatalf("DownloadIUserProfilePic() error = %v", err)
	}
	if string(got) != "fake-profile-pic-bytes" {
		t.Errorf("got %q, want %q", got, "fake-profile-pic-bytes")
	}
}
