package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HttpClient struct {
	Token     string
	UserAgent string
}

func (client HttpClient) setHeaders(req *http.Request) {
	if client.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.Token))
	}
	if client.UserAgent != "" {
		req.Header.Set("User-Agent", client.UserAgent)
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,fa;q=0.8")
	req.Header.Set("Connection", "keep-alive")
}

func (client HttpClient) Get(url string) (string, error) {
	body, err := client.GetBytes(url)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (client HttpClient) GetBytes(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	client.setHeaders(req)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.filimo.com/")

	httpClient := &http.Client{Timeout: 20 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		detail := strings.TrimSpace(string(snippet))
		if d := filimoErrorDetail(detail); d != "" {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, d)
		}
		if detail != "" && len(detail) < 200 {
			return nil, fmt.Errorf("HTTP %d for %s (%s)", resp.StatusCode, url, detail)
		}
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return body, nil
}

func (client HttpClient) Post(url string, payload []byte) (string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	client.setHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	return string(body), nil
}

// DownloadFile دانلود با retry - بدون print تعداد تلاش (stream/download.go این کار رو میکنه)
func (client HttpClient) DownloadFile(url string, filePath string) error {
	return DownloadFileWithResume(url, filePath, client.Token)
}

func filimoErrorDetail(body string) string {
	var payload struct {
		Errors []struct {
			Detail string `json:"detail"`
			Status int    `json:"status"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return ""
	}
	if len(payload.Errors) == 0 {
		return ""
	}
	return strings.TrimSpace(payload.Errors[0].Detail)
}
