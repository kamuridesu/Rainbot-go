package media

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func withCobaltConfig(t *testing.T, url, apiKey string) {
	t.Helper()
	origURL, origKey := CobaltUrl, CobaltApiKey
	CobaltUrl, CobaltApiKey = url, apiKey
	t.Cleanup(func() {
		CobaltUrl, CobaltApiKey = origURL, origKey
	})
}

func TestIsCobaltAvailable(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		key     string
		wantErr error
	}{
		{"missing key", "http://example.com", "", ErrInvalidApiKey},
		{"missing url", "", "key", ErrInvalidUrl},
		{"both configured", "http://example.com", "key", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withCobaltConfig(t, tt.url, tt.key)
			if err := isCobaltAvailable(); err != tt.wantErr {
				t.Errorf("isCobaltAvailable() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCobaltStreamUrlUnavailable(t *testing.T) {
	withCobaltConfig(t, "", "")
	if _, err := GetCobaltStreamUrl(context.Background(), "https://example.com/video", MediaVideo); err == nil {
		t.Error("expected an error when cobalt isn't configured")
	}
}

func TestGetCobaltStreamUrlInvalidQuality(t *testing.T) {
	withCobaltConfig(t, "http://example.com", "key")
	if _, err := GetCobaltStreamUrl(context.Background(), "https://example.com/video", MediaVideo, 999); err == nil {
		t.Error("expected an error for an invalid quality value")
	}
}

func TestGetCobaltStreamUrlSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Api-Key test-key" {
			t.Errorf("Authorization header = %q, want %q", got, "Api-Key test-key")
		}
		fmt.Fprint(w, `{"status":"tunnel","url":"https://cdn.example.com/file.mp4","filename":"file.mp4"}`)
	}))
	defer server.Close()
	withCobaltConfig(t, server.URL, "test-key")

	resp, err := GetCobaltStreamUrl(context.Background(), "https://example.com/video", MediaVideo, 720)
	if err != nil {
		t.Fatalf("GetCobaltStreamUrl() error = %v", err)
	}
	if resp.Status != StatusTunnel || resp.Url != "https://cdn.example.com/file.mp4" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetCobaltStreamUrlNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	withCobaltConfig(t, server.URL, "test-key")

	if _, err := GetCobaltStreamUrl(context.Background(), "https://example.com/video", MediaVideo); err == nil {
		t.Error("expected an error for a non-200 response")
	}
}

func TestDownloadMediaCobaltStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"error","error":{"context":{"service":"youtube"}}}`)
	}))
	defer server.Close()
	withCobaltConfig(t, server.URL, "test-key")

	if _, err := DownloadMediaCobalt(context.Background(), "https://example.com/video", MediaVideo); err == nil {
		t.Error("expected an error for a cobalt 'error' status")
	}
}

func TestDownloadMediaCobaltStatusPicker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"picker"}`)
	}))
	defer server.Close()
	withCobaltConfig(t, server.URL, "test-key")

	if _, err := DownloadMediaCobalt(context.Background(), "https://example.com/video", MediaVideo); err == nil {
		t.Error("expected an error for a cobalt 'picker' status")
	}
}

func TestDownloadMediaCobaltVideoSuccess(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cobalt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"tunnel","url":"http://%s/file","filename":"video.mp4"}`, r.Host)
	})
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-video-bytes"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	withCobaltConfig(t, server.URL+"/cobalt", "test-key")

	result, err := DownloadMediaCobalt(context.Background(), "https://example.com/video", MediaVideo, 720)
	if err != nil {
		t.Fatalf("DownloadMediaCobalt() error = %v", err)
	}
	if string(result.Blob) != "fake-video-bytes" {
		t.Errorf("Blob = %q, want %q", result.Blob, "fake-video-bytes")
	}
	if result.Filename != "video.mp4" {
		t.Errorf("Filename = %q, want %q", result.Filename, "video.mp4")
	}
}

func TestDownloadMediaCobaltDefaultFilename(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cobalt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"status":"tunnel","url":"http://%s/file"}`, r.Host)
	})
	mux.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-video-bytes"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	withCobaltConfig(t, server.URL+"/cobalt", "test-key")

	result, err := DownloadMediaCobalt(context.Background(), "https://example.com/video", MediaVideo)
	if err != nil {
		t.Fatalf("DownloadMediaCobalt() error = %v", err)
	}
	if result.Filename != "media.mp4" {
		t.Errorf("Filename = %q, want default %q", result.Filename, "media.mp4")
	}
}
