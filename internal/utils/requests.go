package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func SendGETRequest[T any](ctx context.Context, client *http.Client, url string, ds *T, headers map[string]string) error {
	slog.Debug(fmt.Sprintf("Sending GET request to: %s", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to build request to %s: %v", url, err)
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to %s: %v", url, err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read body from %s: %v", url, err)
	}

	if res.StatusCode > 400 {
		return fmt.Errorf("error : status is %d and body is %s", res.StatusCode, string(resBody))
	}

	switch v := any(ds).(type) {
	case *string:
		*v = string(resBody)
	case *[]byte:
		*v = resBody
	default:
		if err := json.Unmarshal(resBody, ds); err != nil {
			return fmt.Errorf("failed to read %s response: %v, status is: %d and body is: \n%s", url, err, res.StatusCode, string(resBody))
		}
	}
	return nil
}
