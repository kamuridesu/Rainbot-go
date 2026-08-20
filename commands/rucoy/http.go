package rucoy

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kamuridesu/rainbot-go/core/messages"
)

func sendRucoyGETWithRetry(m *messages.Message, requestURL string) (string, error) {
	delays := []time.Duration{
		5 * time.Second,
		15 * time.Second,
		30 * time.Second,
		60 * time.Second,
	}

	for attempt := 0; attempt <= len(delays); attempt++ {
		req, err := http.NewRequestWithContext(m.Ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return "", fmt.Errorf("failed to build request to %s: %v", requestURL, err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Rainbot-go Rucoy commands)")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to send request to %s: %v", requestURL, err)
		}

		resBody, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("failed to read body from %s: %v", requestURL, readErr)
		}

		if res.StatusCode == http.StatusTooManyRequests {
			if attempt == len(delays) {
				return "", fmt.Errorf("site do Rucoy limitou muitas requisicoes, tente novamente em alguns minutos")
			}

			time.Sleep(rucoyRetryDelay(res, delays[attempt]))
			continue
		}

		if res.StatusCode >= 400 {
			return "", fmt.Errorf("error : status is %d and body is %s", res.StatusCode, string(resBody))
		}

		return string(resBody), nil
	}

	return "", fmt.Errorf("site do Rucoy limitou muitas requisicoes, tente novamente em alguns minutos")
}

func rucoyRetryDelay(res *http.Response, fallback time.Duration) time.Duration {
	retryAfter := strings.TrimSpace(res.Header.Get("Retry-After"))
	if retryAfter == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(retryAfter)
	if err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	retryAt, err := http.ParseTime(retryAfter)
	if err != nil {
		return fallback
	}

	delay := time.Until(retryAt)
	if delay <= 0 {
		return fallback
	}
	return delay
}
