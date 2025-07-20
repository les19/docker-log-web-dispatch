package logs

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LogSender interface {
	SendLog(logLine []byte) error
}

type HTTPLogSender struct {
	client          *http.Client
	authHeaderName  string
	authHeaderValue string
	url             string
}

func NewHTTPLogSender(url string, name string, value string, timeout time.Duration) *HTTPLogSender {
	return &HTTPLogSender{
		client:          &http.Client{Timeout: timeout},
		url:             url,
		authHeaderName:  name,
		authHeaderValue: value,
	}
}

func (s *HTTPLogSender) SendLog(logLine []byte) error {
	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewBuffer(logLine))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for log: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(s.authHeaderName, s.authHeaderValue)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send log to logger service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("logger service returned non-success status %d, could not read response body: %w", resp.StatusCode, readErr)
		}
		return fmt.Errorf("logger service returned non-success status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}
