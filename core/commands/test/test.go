package test

import (
	"log/slog"
	"os"

	"github.com/kamuridesu/rainbot-go/core/commands"
	"github.com/kamuridesu/rainbot-go/core/messages"
)

func init() {

	commands.NewCommand("test", "general test", "meta", nil, nil, false, false, false, func(m *messages.Message) {
		image := `/home/kamuri/Documents/Dev/personal/Rainbot-go/test.webp`
		slog.Info("Loading image: " + image)
		bytes, err := os.ReadFile(image)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("Replying media")
		_, err = m.SendStickerMessage(bytes, messages.ImageMessage, m.RawEvent.Info.Chat)
		if err != nil {
			slog.Error(err.Error())
		}
	})

	commands.NewCommand("testImage", "test sending images", "meta", nil, nil, false, false, false, func(m *messages.Message) {
		image := `/home/kamuri/Pictures/73be9499a73df99623b17fb817bc2154.jpg`
		slog.Info("Loading image: " + image)
		bytes, err := os.ReadFile(image)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("Replying media")
		_, err = m.ReplyMedia(bytes, "test", messages.ImageMessage)
		if err != nil {
			slog.Error(err.Error())
		}
	})

	commands.NewCommand("testSticker", "test sending sticker", "meta", nil, nil, false, false, false, func(m *messages.Message) {
		image := `/home/kamuri/Pictures/73be9499a73df99623b17fb817bc2154.jpg`
		slog.Info("Loading image: " + image)
		bytes, err := os.ReadFile(image)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("Replying media")
		_, err = m.SendStickerMessage(bytes, messages.ImageMessage, m.RawEvent.Info.Chat)
		if err != nil {
			slog.Error(err.Error())
		}
	})

	commands.NewCommand("testVideo", "test sending videos", "meta", nil, nil, false, false, false, func(m *messages.Message) {
		image := `/home/kamuri/Videos/1_5048913860958885334.mp4`
		slog.Info("Loading image: " + image)
		bytes, err := os.ReadFile(image)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		slog.Info("Replying media")
		_, err = m.ReplyMedia(bytes, "test", messages.VideoMessage)
		if err != nil {
			slog.Error(err.Error())
		}
	})

}
