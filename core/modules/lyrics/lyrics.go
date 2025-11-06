package lyrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

const (
	rootUrl          = "https://www.letras.mus.br"
	searchApiRootUrl = "https://solr.sscdn.co/letras/m1/?callback=LetrasSug&q="
)

type Lyrics struct {
	ctx context.Context
}

type ResultData struct {
	DNS string `json:"dns"`
	URL string `json:"url"`
}

type ResponseData struct {
	Response struct {
		Docs []ResultData `json:"docs"`
	} `json:"response"`
}

func hasClass(n *html.Node, className string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			return slices.Contains(classes, className)
		}
	}
	return false
}

func findLyricDiv(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "lyric-original") {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if node := findLyricDiv(c); node != nil {
			return node
		}
	}
	return nil
}

func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
	} else if n.Type == html.ElementNode && n.Data == "br" {
		sb.WriteString("\n")
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}
}

func findParagraphs(n *html.Node, all_text *[]string) {
	if n.Type == html.ElementNode && n.Data == "p" {
		var sb strings.Builder
		extractText(n, &sb)

		text := strings.TrimSpace(sb.String())
		*all_text = append(*all_text, text+"\n")
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findParagraphs(c, all_text)
	}
}

func (l *Lyrics) Search(term string) (*ResultData, error) {
	url := searchApiRootUrl + url.PathEscape(term)

	req, err := http.NewRequestWithContext(l.ctx, "GET", url, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("Unexpected status code: %d and body %s", res.StatusCode, string(body))
		slog.Error(err.Error())
		return nil, err
	}

	result := strings.TrimPrefix(string(body), "LetrasSug(")
	result = strings.ReplaceAll(result, ")", "")
	var data ResponseData
	err = json.Unmarshal([]byte(result), &data)
	if err != nil {
		slog.Error("Error while parsing the json: " + err.Error())
		return nil, err
	}

	if len(data.Response.Docs) == 0 {
		err = fmt.Errorf("No response found")
		slog.Error(err.Error())
		return nil, err
	}

	return &data.Response.Docs[0], nil
}

func (l *Lyrics) GetLyrics(r *ResultData) (*string, error) {
	if r == nil || r.DNS == "" || r.URL == "" {
		err := fmt.Errorf("result data is empty")
		slog.Error(err.Error())
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/$%s", rootUrl, r.DNS, r.URL)

	req, err := http.NewRequestWithContext(l.ctx, "GET", url, nil)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("invalid status code: %d, body is %s", res.StatusCode, string(body))
		slog.Error(err.Error())
		return nil, err
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	var allText []string
	lyricsNode := findLyricDiv(doc)
	if lyricsNode != nil {
		findParagraphs(lyricsNode, &allText)
	}

	result := strings.Join(allText, "\n")

	return &result, nil

}

func SearchLyrics(ctx context.Context, term string) (*string, error) {
	l := Lyrics{ctx: ctx}

	result, err := l.Search(term)
	if err != nil {
		return nil, err
	}

	return l.GetLyrics(result)
}
