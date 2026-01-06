package media

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
)

type CobaltErrorContext struct {
	Service string `json:"service"`
	Limit   int    `json:"limit"`
}

type CobaltError struct {
	Code    int                `json:"code"`
	Context CobaltErrorContext `json:"context"`
}

type CobaltType = string

var (
	TypePhoto CobaltType = "photo"
	TypeVideo CobaltType = "video"
	TypeGif   CobaltType = "gif"
)

type CobaltStatus = string

var (
	StatusError    CobaltStatus = "error"
	StatusPicker   CobaltStatus = "picker"
	StatusTunnel   CobaltStatus = "tunnel"
	StatusRedirect CobaltStatus = "redirect"
)

type PickerResponse struct {
	Type  CobaltType `json:"type"`
	Url   string     `json:"url"`
	Thumb string     `json:"string"`
}

type CobaltResponse struct {
	Status        CobaltStatus     `json:"status"`
	Url           string           `json:"url"`
	Filename      string           `json:"filename"`
	Audio         string           `json:"audio"`
	AudioFilename string           `json:"audioFilename"`
	Picker        []PickerResponse `json:"picker"`
	Error         CobaltError      `json:"error"`
}

type CobaltResult struct {
	Blob     []byte
	Filename string
}

type CobaltRequestBody struct {
	FilenameStyle string `json:"filenameStyle"`
	Url           string `json:"url"`
	DownloadMode  string `json:"downloadMode"`
	VideoQuality  string `json:"videoQuality"`
}

var (
	CobaltUrl        = os.Getenv("COBALT_URL")
	CobaltApiKey     = os.Getenv("COBALT_API_KEY")
	ErrInvalidApiKey = errors.New("invalid cobalt api key")
	ErrInvalidUrl    = errors.New("invalid cobalt url")
)

func isCobaltAvailable() error {
	if CobaltApiKey == "" {
		return ErrInvalidApiKey
	}
	if CobaltUrl == "" {
		return ErrInvalidUrl
	}
	return nil
}

func GetCobaltStreamUrl(ctx context.Context, mediaUrl string, mediaType MediaType, quality ...int) (*CobaltResponse, error) {
	if err := isCobaltAvailable(); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	qual := 360
	if len(quality) > 0 {
		qual = quality[0]
	}
	if !slices.Contains([]int{144, 240, 360, 480, 720, 1080}, qual) {
		err := fmt.Errorf("invalid media quality")
		slog.Error(err.Error())
		return nil, err
	}

	mType := "auto"
	if mediaType == MediaAudio {
		mType = "audio"
	}

	cobaltRequest := CobaltRequestBody{
		Url:           mediaUrl,
		FilenameStyle: "basic",
		DownloadMode:  mType,
		VideoQuality:  strconv.Itoa(qual),
	}

	reqBody, err := json.Marshal(cobaltRequest)
	if err != nil {
		slog.Error("Error while building cobalt body: " + err.Error())
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", CobaltUrl, bytes.NewReader(reqBody))
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Api-Key "+CobaltApiKey)

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

	var response CobaltResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		err = fmt.Errorf("error processing json: %s", err.Error())
		slog.Error(err.Error())
		return nil, err
	}

	return &response, nil

}

func DownloadMediaCobalt(ctx context.Context, mediaUrl string, mediaType MediaType, quality ...int) (*CobaltResult, error) {
	media, err := GetCobaltStreamUrl(ctx, mediaUrl, mediaType, quality...)
	if err != nil {
		return nil, err
	}
	err = nil
	switch media.Status {
	case StatusError:
		err = fmt.Errorf("cobalt returned an error: %s", media.Error.Context.Service)

	case StatusPicker:
		err = fmt.Errorf("cobalt returneda picker, url is invalid")

	}
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", media.Url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("invalid status code: %d, body is %s", res.StatusCode, string(body))
		slog.Error(err.Error())
		return nil, err
	}

	filename := media.Filename
	if filename == "" {
		ext := "mp4"
		if mediaType == MediaAudio {
			ext = "ogg"
		}
		filename = "media." + ext
	}

	if mediaType == MediaAudio {
		body, err = ConvertAudioToOgg(body)
		if err != nil {
			return nil, err
		}
	}

	return &CobaltResult{
		Blob:     body,
		Filename: filename,
	}, nil

}
