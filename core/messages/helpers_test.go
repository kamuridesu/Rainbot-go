package messages

import (
	"reflect"
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func TestParseMessageContentConversation(t *testing.T) {
	rawMsg := &waE2E.Message{Conversation: proto.String("hello world")}

	msgType, text, mentions, quoted, quotedStanzaID, quotedAuthor := parseMessageContent(rawMsg)

	if msgType != TextMessage {
		t.Errorf("msgType = %v, want TextMessage", msgType)
	}
	if text == nil || *text != "hello world" {
		t.Errorf("text = %v, want %q", text, "hello world")
	}
	if len(mentions) != 0 {
		t.Errorf("mentions = %v, want empty", mentions)
	}
	if quoted != nil {
		t.Errorf("quoted = %v, want nil", quoted)
	}
	if quotedStanzaID != "" {
		t.Errorf("quotedStanzaID = %q, want empty", quotedStanzaID)
	}
	if quotedAuthor != "" {
		t.Errorf("quotedAuthor = %q, want empty", quotedAuthor)
	}
}

func TestParseMessageContentExtendedTextWithQuoteAndMentions(t *testing.T) {
	quotedMsg := &waE2E.Message{Conversation: proto.String("original")}
	rawMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String("reply text"),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      proto.String("stanza-123"),
				Participant:   proto.String("111@lid"),
				QuotedMessage: quotedMsg,
				MentionedJID:  []string{"222@lid"},
			},
		},
	}

	msgType, text, mentions, quoted, quotedStanzaID, quotedAuthor := parseMessageContent(rawMsg)

	if msgType != TextMessage {
		t.Errorf("msgType = %v, want TextMessage", msgType)
	}
	if text == nil || *text != "reply text" {
		t.Errorf("text = %v, want %q", text, "reply text")
	}
	if quoted != quotedMsg {
		t.Errorf("quoted = %v, want the same pointer as quotedMsg", quoted)
	}
	if quotedStanzaID != "stanza-123" {
		t.Errorf("quotedStanzaID = %q, want %q", quotedStanzaID, "stanza-123")
	}
	if quotedAuthor != "111@lid" {
		t.Errorf("quotedAuthor = %q, want %q", quotedAuthor, "111@lid")
	}

	wantMentions := []string{"222@lid", "111@lid"}
	if !reflect.DeepEqual(mentions, wantMentions) {
		t.Errorf("mentions = %v, want %v (mentioned JIDs first, then the replied-to participant)", mentions, wantMentions)
	}
}

func TestParseMessageContentExtendedTextWithoutContextInfo(t *testing.T) {
	rawMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{Text: proto.String("no reply here")},
	}

	msgType, text, mentions, quoted, quotedStanzaID, quotedAuthor := parseMessageContent(rawMsg)

	if msgType != TextMessage || text == nil || *text != "no reply here" {
		t.Errorf("unexpected result: type=%v text=%v", msgType, text)
	}
	if len(mentions) != 0 || quoted != nil || quotedStanzaID != "" || quotedAuthor != "" {
		t.Errorf("expected no quote/mention info, got mentions=%v quoted=%v stanzaID=%q author=%q", mentions, quoted, quotedStanzaID, quotedAuthor)
	}
}

func TestParseMessageContentImageMessage(t *testing.T) {
	rawMsg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption: proto.String("nice pic"),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:     proto.String("stanza-456"),
				Participant:  proto.String("111@lid"),
				MentionedJID: []string{"222@lid"},
			},
		},
	}

	msgType, text, mentions, quoted, quotedStanzaID, quotedAuthor := parseMessageContent(rawMsg)

	if msgType != ImageMessage {
		t.Errorf("msgType = %v, want ImageMessage", msgType)
	}
	if text == nil || *text != "nice pic" {
		t.Errorf("text = %v, want %q", text, "nice pic")
	}
	if quotedStanzaID != "stanza-456" {
		t.Errorf("quotedStanzaID = %q, want %q", quotedStanzaID, "stanza-456")
	}
	if quoted != nil {
		t.Errorf("quoted = %v, want nil for an image message", quoted)
	}
	if quotedAuthor != "" {
		t.Errorf("quotedAuthor = %q, want empty for an image message", quotedAuthor)
	}
	wantMentions := []string{"222@lid", "111@lid"}
	if !reflect.DeepEqual(mentions, wantMentions) {
		t.Errorf("mentions = %v, want %v", mentions, wantMentions)
	}
}

func TestParseMessageContentVideoMessage(t *testing.T) {
	rawMsg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption: proto.String("check this out"),
			ContextInfo: &waE2E.ContextInfo{
				StanzaID: proto.String("stanza-789"),
			},
		},
	}

	msgType, text, _, _, quotedStanzaID, _ := parseMessageContent(rawMsg)

	if msgType != VideoMessage {
		t.Errorf("msgType = %v, want VideoMessage", msgType)
	}
	if text == nil || *text != "check this out" {
		t.Errorf("text = %v, want %q", text, "check this out")
	}
	if quotedStanzaID != "stanza-789" {
		t.Errorf("quotedStanzaID = %q, want %q", quotedStanzaID, "stanza-789")
	}
}

func TestParseMessageContentStickerAudioReactionAndUnknown(t *testing.T) {
	tests := []struct {
		name string
		msg  *waE2E.Message
		want MessageType
	}{
		{"sticker", &waE2E.Message{StickerMessage: &waE2E.StickerMessage{}}, StickerMessage},
		{"audio", &waE2E.Message{AudioMessage: &waE2E.AudioMessage{}}, AudioMessage},
		{"reaction", &waE2E.Message{ReactionMessage: &waE2E.ReactionMessage{}}, ReactionMessage},
		{"unrecognized/empty message", &waE2E.Message{}, Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgType, _, _, _, _, _ := parseMessageContent(tt.msg)
			if msgType != tt.want {
				t.Errorf("msgType = %v, want %v", msgType, tt.want)
			}
		})
	}
}

func TestParseCommandArgs(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		prefix   string
		wantCmd  *string
		wantArgs *[]string
	}{
		{
			name:     "command with args",
			text:     "/quotly foo bar",
			prefix:   "/",
			wantCmd:  strPtr("quotly"),
			wantArgs: strSlicePtr([]string{"foo", "bar"}),
		},
		{
			name:     "command without args",
			text:     "/help",
			prefix:   "/",
			wantCmd:  strPtr("help"),
			wantArgs: strSlicePtr([]string{}),
		},
		{
			name:     "empty text yields no command",
			text:     "",
			prefix:   "/",
			wantCmd:  nil,
			wantArgs: nil,
		},
		{
			name:     "prefix only strips a leading match",
			text:     "hello world",
			prefix:   "/",
			wantCmd:  strPtr("hello"),
			wantArgs: strSlicePtr([]string{"world"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := parseCommandArgs(tt.text, tt.prefix)

			if (cmd == nil) != (tt.wantCmd == nil) {
				t.Fatalf("cmd = %v, want %v", cmd, tt.wantCmd)
			}
			if cmd != nil && *cmd != *tt.wantCmd {
				t.Errorf("cmd = %q, want %q", *cmd, *tt.wantCmd)
			}

			if (args == nil) != (tt.wantArgs == nil) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			if args != nil && !reflect.DeepEqual(*args, *tt.wantArgs) {
				t.Errorf("args = %v, want %v", *args, *tt.wantArgs)
			}
		})
	}
}

func strPtr(s string) *string { return &s }

func strSlicePtr(s []string) *[]string { return &s }

func TestSafeText(t *testing.T) {
	text := "hello"
	m := &Message{Text: &text}
	if got := m.safeText(); got != "hello" {
		t.Errorf("safeText() = %q, want %q", got, "hello")
	}

	m = &Message{Text: nil}
	if got := m.safeText(); got != "" {
		t.Errorf("safeText() with nil Text = %q, want empty string", got)
	}
}

func TestGetAuthorJID(t *testing.T) {
	t.Run("sender is already @lid", func(t *testing.T) {
		v := &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Sender: types.NewJID("111", types.HiddenUserServer),
				},
			},
		}
		got := getAuthorJID(v)
		if got != "111@lid" {
			t.Errorf("getAuthorJID() = %q, want %q", got, "111@lid")
		}
	})

	t.Run("falls back to SenderAlt when Sender isn't @lid", func(t *testing.T) {
		v := &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Sender:    types.NewJID("5511999999999", types.DefaultUserServer),
					SenderAlt: types.NewJID("222", types.HiddenUserServer),
				},
			},
		}
		got := getAuthorJID(v)
		if got != "222@lid" {
			t.Errorf("getAuthorJID() = %q, want %q", got, "222@lid")
		}
	})

	t.Run("panics when neither Sender nor SenderAlt is @lid", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("expected getAuthorJID to panic when no @lid address is available")
			}
		}()
		v := &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Sender:    types.NewJID("5511999999999", types.DefaultUserServer),
					SenderAlt: types.NewJID("5511888888888", types.DefaultUserServer),
				},
			},
		}
		getAuthorJID(v)
	})
}
