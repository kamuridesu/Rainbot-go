package utils

import (
	"context"
	"io"
	"net/http"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func DownloadIUserProfilePic(ctx context.Context, jid_ string, bot *bot.Bot) ([]byte, error) {
	jid, err := types.ParseJID(jid_)
	if err != nil {
		return nil, err
	}
	pp, err := bot.Client.GetProfilePictureInfo(ctx, jid, &whatsmeow.GetProfilePictureParams{})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", pp.URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(res.Body)
}
