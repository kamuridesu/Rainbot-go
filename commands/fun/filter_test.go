package fun

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/database/models"
	"github.com/kamuridesu/rainbot-go/core/messages"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

func TestNewFilterRejectsMissingQuote(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.QuotedMessage = nil
	*m.Args = []string{"oi"}

	NewFilter(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + rejection), got %d", len(fake.SentMessages))
	}
}

func TestNewFilterRejectsUnsupportedQuoteType(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	m.QuotedMessage = &messages.Message{Type: messages.AudioMessage}
	*m.Args = []string{"oi"}

	NewFilter(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (reaction + rejection), got %d", len(fake.SentMessages))
	}
}

func TestNewFilterRejectsInvalidRegex(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	text := "hello"
	m.QuotedMessage = &messages.Message{Type: messages.TextMessage, Text: &text}
	*m.Args = []string{"("}

	NewFilter(m)

	last := fake.SentMessages[len(fake.SentMessages)-1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(last, "invalido") {
		t.Errorf("expected an invalid-regex reply, got %q", last)
	}
}

func TestNewFilterRejectsDuplicatePattern(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: "text", Response: "ola"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	text := "hello"
	m.QuotedMessage = &messages.Message{Type: messages.TextMessage, Text: &text}
	*m.Args = []string{"oi"}

	NewFilter(m)

	last := fake.SentMessages[len(fake.SentMessages)-1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(last, "já existe") {
		t.Errorf("expected a duplicate-pattern reply, got %q", last)
	}
}

func TestNewFilterTextSuccess(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)

	quotedText := "resposta automatica"
	m.QuotedMessage = &messages.Message{Type: messages.TextMessage, Text: &quotedText}
	*m.Args = []string{"oi"}

	NewFilter(m)

	filters, err := db.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		t.Fatalf("GetFilters() error = %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 saved filter, got %d", len(filters))
	}
	if filters[0].Kind != "text" || filters[0].Response != quotedText {
		t.Errorf("unexpected saved filter: %+v", filters[0])
	}
}

func TestNewFilterImageSuccess(t *testing.T) {
	db := newTestDB(t)

	t.Chdir(t.TempDir())

	fake := &botfakes.FakeClient{
		DownloadFunc: func(ctx context.Context, msg whatsmeow.DownloadableMessage) ([]byte, error) {
			return []byte("image-bytes"), nil
		},
	}
	m := newTestMessage(t, fake, db)
	emptyCaption := ""
	m.QuotedMessage = &messages.Message{
		Type:       messages.ImageMessage,
		Text:       &emptyCaption,
		RawMessage: &waE2E.Message{ImageMessage: &waE2E.ImageMessage{}},
	}
	*m.Args = []string{"figura"}

	NewFilter(m)

	filters, err := db.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		t.Fatalf("GetFilters() error = %v", err)
	}
	if len(filters) != 1 {
		t.Fatalf("expected 1 saved filter, got %d", len(filters))
	}
	if filters[0].Kind != "image" {
		t.Errorf("Kind = %q, want %q", filters[0].Kind, "image")
	}
	if !strings.HasSuffix(filters[0].Response, ".jpg") {
		t.Errorf("expected the response to be a saved .jpg filename, got %q", filters[0].Response)
	}
	if _, err := filepath.Abs(filters[0].Response); err != nil {
		t.Errorf("unexpected path error: %v", err)
	}
}

func TestShowFiltersEmpty(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))

	ShowFilters(m)

	last := fake.SentMessages[len(fake.SentMessages)-1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(last, "Nenhum filtro") {
		t.Errorf("expected an empty-filters reply, got %q", last)
	}
}

func TestShowFiltersListsPatterns(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: "text", Response: "ola"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}

	ShowFilters(m)

	last := fake.SentMessages[len(fake.SentMessages)-1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(last, "oi") {
		t.Errorf("expected the filter pattern to be listed, got %q", last)
	}
}

func TestDeleteFilterNotFound(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"nao-existe"}

	DeleteFilter(m)

	last := fake.SentMessages[len(fake.SentMessages)-1].Message.GetExtendedTextMessage().GetText()
	if !strings.Contains(last, "No filter") {
		t.Errorf("expected a not-found reply, got %q", last)
	}
}

func TestDeleteFilterSuccess(t *testing.T) {
	db := newTestDB(t)
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, db)
	if err := db.Filter.NewFilter(&models.Filter{ChatID: m.Chat.ChatID, Pattern: "oi", Kind: "text", Response: "ola"}); err != nil {
		t.Fatalf("seed error = %v", err)
	}
	*m.Args = []string{"oi"}

	DeleteFilter(m)

	if len(fake.SentMessages) != 1 {
		t.Fatalf("expected 1 sent message (success reaction), got %d", len(fake.SentMessages))
	}
	filters, err := db.Filter.GetFilters(m.Chat.ChatID)
	if err != nil {
		t.Fatalf("GetFilters() error = %v", err)
	}
	if len(filters) != 0 {
		t.Errorf("expected the filter to be gone, got %+v", filters)
	}
}
