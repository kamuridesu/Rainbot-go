package lyrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func mustParseHTML(t *testing.T, s string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatalf("failed to parse test HTML: %v", err)
	}
	return doc
}

func TestHasClass(t *testing.T) {
	doc := mustParseHTML(t, `<div class="foo lyric-original bar"></div>`)
	div := findFirstElement(doc, "div")
	if div == nil {
		t.Fatal("test setup: div not found")
	}
	if !hasClass(div, "lyric-original") {
		t.Error("expected hasClass to find 'lyric-original' among multiple classes")
	}
	if hasClass(div, "does-not-exist") {
		t.Error("expected hasClass to return false for an absent class")
	}
}

func findFirstElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

func TestFindLyricDiv(t *testing.T) {
	doc := mustParseHTML(t, `<html><body><div class="other"></div><div class="lyric-original">hi</div></body></html>`)
	node := findLyricDiv(doc)
	if node == nil {
		t.Fatal("expected to find the lyric-original div")
	}
	if !hasClass(node, "lyric-original") {
		t.Error("found node doesn't have the lyric-original class")
	}
}

func TestFindLyricDivNotFound(t *testing.T) {
	doc := mustParseHTML(t, `<html><body><div class="other"></div></body></html>`)
	if findLyricDiv(doc) != nil {
		t.Error("expected nil when no lyric-original div exists")
	}
}

func TestExtractText(t *testing.T) {
	doc := mustParseHTML(t, `<p>hello<br>world</p>`)
	p := findFirstElement(doc, "p")
	var sb strings.Builder
	extractText(p, &sb)
	if sb.String() != "hello\nworld" {
		t.Errorf("extractText() = %q, want %q", sb.String(), "hello\nworld")
	}
}

func TestFindParagraphs(t *testing.T) {
	doc := mustParseHTML(t, `<div><p>first line</p><p>second<br>line</p></div>`)
	var paragraphs []string
	findParagraphs(doc, &paragraphs)

	if len(paragraphs) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d: %v", len(paragraphs), paragraphs)
	}
	if paragraphs[0] != "first line\n" {
		t.Errorf("paragraphs[0] = %q, want %q", paragraphs[0], "first line\n")
	}
	if paragraphs[1] != "second\nline\n" {
		t.Errorf("paragraphs[1] = %q, want %q", paragraphs[1], "second\nline\n")
	}
}

func TestSearchSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `LetrasSug({"response":{"docs":[{"dns":"artist","url":"song"}]}})`)
	}))
	defer server.Close()

	origURL := searchApiRootUrl
	searchApiRootUrl = server.URL + "/?callback=LetrasSug&q="
	t.Cleanup(func() { searchApiRootUrl = origURL })

	l := &Lyrics{ctx: context.Background()}
	result, err := l.Search("some song")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if result.DNS != "artist" || result.URL != "song" {
		t.Errorf("Search() = %+v, want DNS=artist URL=song", result)
	}
}

func TestSearchNoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `LetrasSug({"response":{"docs":[]}})`)
	}))
	defer server.Close()

	origURL := searchApiRootUrl
	searchApiRootUrl = server.URL + "/?callback=LetrasSug&q="
	t.Cleanup(func() { searchApiRootUrl = origURL })

	l := &Lyrics{ctx: context.Background()}
	_, err := l.Search("some song")
	if err == nil {
		t.Fatal("expected an error when no results are found")
	}
}

func TestSearchNon200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	origURL := searchApiRootUrl
	searchApiRootUrl = server.URL + "/?callback=LetrasSug&q="
	t.Cleanup(func() { searchApiRootUrl = origURL })

	l := &Lyrics{ctx: context.Background()}
	_, err := l.Search("some song")
	if err == nil {
		t.Fatal("expected an error for a non-200 response")
	}
}

func TestGetLyricsNilResult(t *testing.T) {
	l := &Lyrics{ctx: context.Background()}
	if _, err := l.GetLyrics(nil); err == nil {
		t.Error("expected an error for a nil result")
	}
	if _, err := l.GetLyrics(&ResultData{}); err == nil {
		t.Error("expected an error for an empty result")
	}
}

func TestGetLyricsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><div class="lyric-original"><p>line one</p><p>line two</p></div></body></html>`)
	}))
	defer server.Close()

	origURL := rootUrl
	rootUrl = server.URL
	t.Cleanup(func() { rootUrl = origURL })

	l := &Lyrics{ctx: context.Background()}
	got, err := l.GetLyrics(&ResultData{DNS: "artist", URL: "song"})
	if err != nil {
		t.Fatalf("GetLyrics() error = %v", err)
	}
	if !strings.Contains(*got, "line one") || !strings.Contains(*got, "line two") {
		t.Errorf("GetLyrics() = %q, want it to contain both lines", *got)
	}
}

func TestSearchLyricsEndToEnd(t *testing.T) {
	searchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `LetrasSug({"response":{"docs":[{"dns":"artist","url":"song"}]}})`)
	}))
	defer searchServer.Close()

	lyricsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><div class="lyric-original"><p>never gonna give you up</p></div></body></html>`)
	}))
	defer lyricsServer.Close()

	origSearch, origRoot := searchApiRootUrl, rootUrl
	searchApiRootUrl = searchServer.URL + "/?callback=LetrasSug&q="
	rootUrl = lyricsServer.URL
	t.Cleanup(func() {
		searchApiRootUrl = origSearch
		rootUrl = origRoot
	})

	got, err := SearchLyrics(context.Background(), "never gonna give you up")
	if err != nil {
		t.Fatalf("SearchLyrics() error = %v", err)
	}
	if !strings.Contains(*got, "never gonna give you up") {
		t.Errorf("SearchLyrics() = %q, want it to contain the lyric line", *got)
	}
}
