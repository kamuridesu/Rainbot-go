package fun

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
	"github.com/kamuridesu/rainbot-go/internal/configs"
)

func withCobaltServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	withCobaltConfig(t, server.URL, "test-key")
}

func withCobaltConfig(t *testing.T, url, apiKey string) {
	t.Helper()
	configs.Override(t, func(c *configs.Config) {
		c.Cobalt = configs.Cobalt{Url: url, ApiKey: apiKey}
	})
}

func TestDownloadVideoSuccess(t *testing.T) {
	withCobaltServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/file" {
			w.Write([]byte("fake-video-bytes"))
			return
		}
		fmt.Fprintf(w, `{"status":"tunnel","url":"http://%s/file","filename":"video.mp4"}`, r.Host)
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"https://example.com/video"}

	DownloadVideo(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + success reaction + video), got %d", len(fake.SentMessages))
	}
	if fake.SentMessages[2].Message.GetVideoMessage() == nil {
		t.Error("expected a video message to be sent")
	}
}

func TestDownloadVideoError(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"https://example.com/video"}

	withCobaltConfig(t, "", "")

	DownloadVideo(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + fail reaction + error reply), got %d", len(fake.SentMessages))
	}
}

func TestDownloadAudioNotImplemented(t *testing.T) {
	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"https://example.com/audio"}

	DownloadAudio(m)

	if len(fake.SentMessages) != 2 {
		t.Fatalf("expected 2 sent messages (fail reaction + not implemented reply), got %d", len(fake.SentMessages))
	}
	if got := fake.SentMessages[1].Message.GetExtendedTextMessage().GetText(); got != "Não implementado" {
		t.Errorf("reply = %q, want %q", got, "Não implementado")
	}
}
