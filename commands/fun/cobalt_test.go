package fun

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kamuridesu/rainbot-go/core/modules/media"
	"github.com/kamuridesu/rainbot-go/internal/bot/botfakes"
)

func withCobaltServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	origURL, origKey := media.CobaltUrl, media.CobaltApiKey
	media.CobaltUrl, media.CobaltApiKey = server.URL, "test-key"
	t.Cleanup(func() {
		media.CobaltUrl, media.CobaltApiKey = origURL, origKey
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

	media.CobaltUrl, media.CobaltApiKey = "", ""

	DownloadVideo(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + fail reaction + error reply), got %d", len(fake.SentMessages))
	}
}

func TestDownloadAudioConversionError(t *testing.T) {
	withCobaltServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/file" {
			w.Write([]byte("not-real-audio-bytes"))
			return
		}
		fmt.Fprintf(w, `{"status":"tunnel","url":"http://%s/file","filename":"audio.ogg"}`, r.Host)
	})

	fake := &botfakes.FakeClient{}
	m := newTestMessage(t, fake, newTestDB(t))
	*m.Args = []string{"https://example.com/audio"}

	DownloadAudio(m)

	if len(fake.SentMessages) != 3 {
		t.Fatalf("expected 3 sent messages (waiting + fail reaction + error reply), got %d", len(fake.SentMessages))
	}
}
