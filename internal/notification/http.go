package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/intraware/rodan/internal/config"
)

type notifPayload struct {
	Message string `json:"message"`
}

func httpNotif(message string, cfg config.NotificationConfig) error {
	if cfg.HTTP == nil {
		return fmt.Errorf("HTTP notification config is nil")
	}
	url := cfg.HTTP.URL
	endpoint := cfg.HTTP.Endpoint
	apiKey := cfg.HTTP.HashedAPIKey
	body, err := json.Marshal(notifPayload{Message: message})
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	retries := cfg.DefaultRetry
	for i := range retries {
		req, err := http.NewRequest("POST", url+endpoint, bytes.NewBuffer(body))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		client := &http.Client{Timeout: cfg.Timeout}
		resp, err := client.Do(req)
		if err != nil {
			if i < retries {
				time.Sleep(cfg.RetryDelay)
				continue
			}
			return fmt.Errorf("failed to send notification: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			if i < retries {
				time.Sleep(cfg.RetryDelay)
				continue
			}
			return fmt.Errorf("notification failed with status: %s", resp.Status)
		}
		return nil
	}
	return fmt.Errorf("notification failed after %d retries", retries)
}
