package quotly

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

func Generate(ctx context.Context, body_ QuotlyRequestBody) ([]byte, error) {
	quotlyApiUrl := os.Getenv("QUOTLY_API_URL")
	if quotlyApiUrl == "" {
		return nil, errors.New("Variable `QUOTLY_API_URL` not set")
	}
	data, err := json.Marshal(body_)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/generate", quotlyApiUrl), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		slog.Info("Failed to generate: " + string(body))
		return nil, errors.New("Failed to generate: " + string(body))
	}

	var response QuotlyResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(response.Result.Image)
}
