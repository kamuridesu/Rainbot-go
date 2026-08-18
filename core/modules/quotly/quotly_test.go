package quotly

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateMissingApiUrl(t *testing.T) {
	t.Setenv("QUOTLY_API_URL", "")

	if _, err := Generate(context.Background(), DefaultTemplate); err == nil {
		t.Error("expected an error when QUOTLY_API_URL is unset")
	}
}

func TestGenerateSuccess(t *testing.T) {
	imageBytes := []byte("fake-png-bytes")
	encoded := base64.StdEncoding.EncodeToString(imageBytes)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/generate" {
			t.Errorf("request path = %q, want %q", r.URL.Path, "/generate")
		}
		fmt.Fprintf(w, `{"ok":true,"result":{"image":%q,"width":512,"height":768}}`, encoded)
	}))
	defer server.Close()
	t.Setenv("QUOTLY_API_URL", server.URL)

	got, err := Generate(context.Background(), DefaultTemplate)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if string(got) != string(imageBytes) {
		t.Errorf("Generate() = %q, want %q", got, imageBytes)
	}
}

func TestGenerateNon200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"bad request"}`)
	}))
	defer server.Close()
	t.Setenv("QUOTLY_API_URL", server.URL)

	if _, err := Generate(context.Background(), DefaultTemplate); err == nil {
		t.Error("expected an error for a non-200 response")
	}
}

func TestGenerateInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	defer server.Close()
	t.Setenv("QUOTLY_API_URL", server.URL)

	if _, err := Generate(context.Background(), DefaultTemplate); err == nil {
		t.Error("expected an error for an invalid JSON response")
	}
}
