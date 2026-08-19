package messages

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"github.com/kamuridesu/rainbot-go/internal/emojis"
	"github.com/kamuridesu/rainbot-go/internal/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func newTestMessage(fake *botfakes.FakeClient) *Message {
	chatJID := types.NewJID("123456", types.GroupServer)
	senderJID := types.NewJID("111", types.HiddenUserServer)

	return &Message{
		Ctx: context.Background(),
		Bot: &bot.Bot{Client: fake},
		RawEvent: &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Chat:     chatJID,
					Sender:   senderJID,
					IsGroup:  true,
					IsFromMe: false,
				},
				ID: "orig-stanza-id",
			},
		},
	}
}

func TestReplyRejectsEmptyContent(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	_, err := m.Reply("   ")
	if err == nil {
		t.Fatal("expected an error for empty/whitespace-only content")
	}
	if len(fake.SentMessages) != 0 {
		t.Errorf("expected no message to be sent, got %v", fake.SentMessages)
	}
}

func TestReplySendsExtendedTextMessage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.Reply("hello world"); err != nil {
		t.Fatalf("Reply() error = %v", err)
	}

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(fake.SentMessages))
	}
	sent := fake.SentMessages[0]
	if sent.Message.GetExtendedTextMessage().GetText() != "hello world" {
		t.Errorf("sent text = %q, want %q", sent.Message.GetExtendedTextMessage().GetText(), "hello world")
	}
	if sent.Chat != m.RawEvent.Info.Chat {
		t.Errorf("sent to chat %v, want %v", sent.Chat, m.RawEvent.Info.Chat)
	}
}

func TestReplyWithReactionSendsBoth(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.Reply("hello", emojis.Success); err != nil {
		t.Fatalf("Reply() error = %v", err)
	}

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 messages sent (reaction + reply), got %d", len(fake.SentMessages))
	}
}

func TestReplyPropagatesReactionError(t *testing.T) {
	wantErr := errors.New("send failed")
	fake := &botfakes.FakeClient{
		SendMessageFunc: func(ctx context.Context, to types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
			return whatsmeow.SendResponse{}, wantErr
		},
	}
	m := newTestMessage(fake)

	_, err := m.Reply("hello", emojis.Success)
	if !errors.Is(err, wantErr) {
		t.Errorf("Reply() with a failing reaction, error = %v, want %v", err, wantErr)
	}
}

func TestReplyMentionsAreRewritten(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.Reply("hey 111@lid how are you"); err != nil {
		t.Fatalf("Reply() error = %v", err)
	}

	sent := fake.SentMessages[0]
	gotText := sent.Message.GetExtendedTextMessage().GetText()
	if gotText != "hey @111 how are you" {
		t.Errorf("sent text = %q, want mention rewritten to %q", gotText, "hey @111 how are you")
	}
	gotMentions := sent.Message.GetExtendedTextMessage().GetContextInfo().GetMentionedJID()
	if len(gotMentions) != 1 || gotMentions[0] != "111@lid" {
		t.Errorf("mentioned JIDs = %v, want [111@lid]", gotMentions)
	}
}

func TestReact(t *testing.T) {
	fake := &botfakes.FakeClient{}
	var gotChat, gotSender types.JID
	var gotID types.MessageID
	var gotReaction string
	fake.BuildReactionFunc = func(chat, sender types.JID, id types.MessageID, reaction string) *waE2E.Message {
		gotChat, gotSender, gotID, gotReaction = chat, sender, id, reaction
		return &waE2E.Message{}
	}
	m := newTestMessage(fake)

	if _, err := m.React(emojis.Success); err != nil {
		t.Fatalf("React() error = %v", err)
	}
	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(fake.SentMessages))
	}
	if gotChat != m.RawEvent.Info.Chat || gotSender != m.RawEvent.Info.Sender || gotID != m.RawEvent.Info.ID {
		t.Errorf("BuildReaction() called with (%v, %v, %v), want (%v, %v, %v)",
			gotChat, gotSender, gotID, m.RawEvent.Info.Chat, m.RawEvent.Info.Sender, m.RawEvent.Info.ID)
	}
	if gotReaction != string(emojis.Success) {
		t.Errorf("reaction = %q, want %q", gotReaction, string(emojis.Success))
	}
}

func TestDelete(t *testing.T) {
	fake := &botfakes.FakeClient{}
	buildRevokeCalled := false
	fake.BuildRevokeFunc = func(chat, sender types.JID, id types.MessageID) *waE2E.Message {
		buildRevokeCalled = true
		return &waE2E.Message{}
	}
	m := newTestMessage(fake)

	if _, err := m.Delete(); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if !buildRevokeCalled {
		t.Error("expected BuildRevoke() to be called")
	}
	if len(fake.SentMessages) != 1 {
		t.Errorf("expected 1 message sent, got %d", len(fake.SentMessages))
	}
}

func TestEdit(t *testing.T) {
	fake := &botfakes.FakeClient{}
	var gotContent string
	fake.BuildEditFunc = func(chat types.JID, id types.MessageID, newContent *waE2E.Message) *waE2E.Message {
		gotContent = newContent.GetConversation()
		return &waE2E.Message{}
	}
	m := newTestMessage(fake)

	if _, err := m.Edit("new text"); err != nil {
		t.Fatalf("Edit() error = %v", err)
	}
	if gotContent != "new text" {
		t.Errorf("BuildEdit() received content %q, want %q", gotContent, "new text")
	}
}

func TestIsFromGroup(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if !m.IsFromGroup() {
		t.Error("expected IsFromGroup() to be true")
	}

	m.RawEvent.Info.IsGroup = false
	if m.IsFromGroup() {
		t.Error("expected IsFromGroup() to be false")
	}
}

func TestSendMessage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	msg := &waE2E.Message{Conversation: proto.String("direct send")}
	chatID := types.NewJID("999", types.GroupServer)

	if _, err := m.SendMessage(msg, chatID); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if len(fake.SentMessages) != 1 || fake.SentMessages[0].Chat != chatID {
		t.Errorf("expected message sent to %v, got %v", chatID, fake.SentMessages)
	}
}

func TestSendImageMessage(t *testing.T) {
	fake := &botfakes.FakeClient{
		UploadFunc: func(ctx context.Context, plaintext []byte, appInfo whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
			return whatsmeow.UploadResponse{URL: "https://example.com/img"}, nil
		},
	}
	m := newTestMessage(fake)

	chatID := types.NewJID("999", types.GroupServer)
	if _, err := m.SendImageMessage("a caption", []byte("fake-image-bytes"), chatID); err != nil {
		t.Fatalf("SendImageMessage() error = %v", err)
	}
	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(fake.SentMessages))
	}
	img := fake.SentMessages[0].Message.GetImageMessage()
	if img.GetCaption() != "a caption" {
		t.Errorf("caption = %q, want %q", img.GetCaption(), "a caption")
	}
	if img.GetURL() != "https://example.com/img" {
		t.Errorf("URL = %q, want the uploaded URL", img.GetURL())
	}
}

func TestSendImageMessageUploadError(t *testing.T) {
	wantErr := errors.New("upload failed")
	fake := &botfakes.FakeClient{
		UploadFunc: func(ctx context.Context, plaintext []byte, appInfo whatsmeow.MediaType) (whatsmeow.UploadResponse, error) {
			return whatsmeow.UploadResponse{}, wantErr
		},
	}
	m := newTestMessage(fake)

	_, err := m.SendImageMessage("caption", []byte("data"), types.NewJID("999", types.GroupServer))
	if !errors.Is(err, wantErr) {
		t.Errorf("SendImageMessage() error = %v, want %v", err, wantErr)
	}
	if len(fake.SentMessages) != 0 {
		t.Error("expected no message to be sent when upload fails")
	}
}

func TestSendVideoMessage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.SendVideoMessage("caption", []byte("video-bytes"), types.NewJID("999", types.GroupServer)); err != nil {
		t.Fatalf("SendVideoMessage() error = %v", err)
	}
	if len(fake.SentMessages) != 1 || fake.SentMessages[0].Message.GetVideoMessage() == nil {
		t.Errorf("expected a video message to be sent, got %v", fake.SentMessages)
	}
}

func TestSendAudioMessage(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.SendAudioMessage([]byte("audio-bytes"), types.NewJID("999", types.GroupServer)); err != nil {
		t.Fatalf("SendAudioMessage() error = %v", err)
	}
	if len(fake.SentMessages) != 1 || fake.SentMessages[0].Message.GetAudioMessage() == nil {
		t.Errorf("expected an audio message to be sent, got %v", fake.SentMessages)
	}
}

func TestSendMediaMessageDispatchesByType(t *testing.T) {
	tests := []struct {
		check   func(msg *waE2E.Message) bool
		name    string
		msgType MessageType
	}{
		{"image", ImageMessage, func(msg *waE2E.Message) bool { return msg.GetImageMessage() != nil }},
		{"video", VideoMessage, func(msg *waE2E.Message) bool { return msg.GetVideoMessage() != nil }},
		{"audio", AudioMessage, func(msg *waE2E.Message) bool { return msg.GetAudioMessage() != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &botfakes.FakeClient{}
			m := newTestMessage(fake)

			if _, err := m.SendMediaMessage([]byte("data"), "caption", tt.msgType, types.NewJID("999", types.GroupServer)); err != nil {
				t.Fatalf("SendMediaMessage() error = %v", err)
			}
			if len(fake.SentMessages) != 1 || !tt.check(fake.SentMessages[0].Message) {
				t.Errorf("SendMediaMessage(%v) didn't send the expected message type", tt.msgType)
			}
		})
	}
}

func TestSendMediaMessageUnsupportedType(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	_, err := m.SendMediaMessage([]byte("data"), "caption", StickerMessage, types.NewJID("999", types.GroupServer))
	if err == nil {
		t.Fatal("expected an error for an unsupported media type")
	}
	if len(fake.SentMessages) != 0 {
		t.Error("expected no message to be sent for an unsupported media type")
	}
}

func TestHasValidMedia(t *testing.T) {
	tests := []struct {
		name          string
		msgType       MessageType
		ignoreSticker bool
		want          bool
	}{
		{"image is valid", ImageMessage, false, true},
		{"video is valid", VideoMessage, false, true},
		{"sticker is valid by default", StickerMessage, false, true},
		{"sticker invalid when ignored", StickerMessage, true, false},
		{"text is never valid", TextMessage, false, false},
		{"image still valid when ignoring stickers", ImageMessage, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{Type: tt.msgType}
			got := m.HasValidMedia(tt.ignoreSticker)
			if got != tt.want {
				t.Errorf("HasValidMedia(ignoreSticker=%v) with type %v = %v, want %v", tt.ignoreSticker, tt.msgType, got, tt.want)
			}
		})
	}
}

func TestHasValidMediaNoArgsDefaultsToIncludingSticker(t *testing.T) {
	m := &Message{Type: StickerMessage}
	if !m.HasValidMedia() {
		t.Error("expected sticker to be valid media when ignoreSticker isn't passed")
	}
}

func TestReplyMedia(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.ReplyMedia([]byte("data"), "a caption", ImageMessage, emojis.Success); err != nil {
		t.Fatalf("ReplyMedia() error = %v", err)
	}

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + media), got %d", len(fake.SentMessages))
	}
	img := fake.SentMessages[1].Message.GetImageMessage()
	if img == nil || img.GetCaption() != "a caption" {
		t.Errorf("expected an image message with the given caption, got %+v", fake.SentMessages[1].Message)
	}
}

func TestReplyMediaReactionError(t *testing.T) {
	wantErr := errors.New("react failed")
	fake := &botfakes.FakeClient{
		SendMessageFunc: func(ctx context.Context, to types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
			return whatsmeow.SendResponse{}, wantErr
		},
	}
	m := newTestMessage(fake)

	if _, err := m.ReplyMedia([]byte("data"), "caption", ImageMessage, emojis.Fail); !errors.Is(err, wantErr) {
		t.Errorf("ReplyMedia() error = %v, want %v", err, wantErr)
	}
}

func testWebPBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test PNG: %v", err)
	}
	webpBytes, err := utils.ToWebp(buf.Bytes())
	if err != nil {
		t.Fatalf("failed to convert test image to webp: %v", err)
	}
	return webpBytes
}

func TestReplySticker(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.ReplySticker(testWebPBytes(t), ImageMessage, emojis.Success); err != nil {
		t.Fatalf("ReplySticker() error = %v", err)
	}

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + sticker), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[1].Message.GetStickerMessage() == nil {
		t.Error("expected a sticker message to be sent")
	}
}

func TestReplyStickerNoReaction(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(fake)

	if _, err := m.ReplySticker(testWebPBytes(t), ImageMessage); err != nil {
		t.Fatalf("ReplySticker() error = %v", err)
	}

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (no reaction requested), got %d", len(fake.SentMessages))
	}
}
