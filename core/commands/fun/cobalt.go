package fun

import (
	"strings"

	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/core/modules/media"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
)

func DownloadVideo(m *messages.Message) {
	args := strings.Join(*m.Args, " ")

	m.React(emojis.Waiting)
	res, err := media.DownloadMediaCobalt(m.Ctx, args, media.MediaVideo, 720)
	if err != nil {
		m.Reply("Erro ao baixar: "+err.Error(), emojis.Fail)
		return
	}

	m.ReplyMedia(res.Blob, res.Filename, messages.VideoMessage, emojis.Success)
}

func DownloadAudio(m *messages.Message) {
	args := strings.Join(*m.Args, " ")
	m.React(emojis.Waiting)
	res, err := media.DownloadMediaCobalt(m.Ctx, args, media.MediaAudio, 360)
	if err != nil {
		m.Reply("Erro ao baixar: "+err.Error(), emojis.Fail)
		return
	}

	m.ReplyMedia(res.Blob, res.Filename, messages.AudioMessage, emojis.Success)
}
