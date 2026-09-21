package helpers

import (
	"errors"
	"io"
	"net/http"
	"time"
)

// FetchResult — результат скачивания URL.
type FetchResult struct {
	Gz      []byte
	Mime    string
	RawSize int
}

// FetchContent скачивает URL, сжимает ответ gzip и возвращает результат.
func FetchContent(url string, timeout time.Duration, maxBytes int64) (*FetchResult, error) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "altaridge-eedg-sl/1.0")
	req.Header.Set("Accept-Encoding", "identity") // сами жмём — не хотим двойного

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.New("upstream returned " + resp.Status)
	}

	limited := io.LimitReader(resp.Body, maxBytes)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	gz, err := GzipBytes(raw)
	if err != nil {
		return nil, err
	}

	return &FetchResult{
		Gz:      gz,
		Mime:    mime,
		RawSize: len(raw),
	}, nil
}
