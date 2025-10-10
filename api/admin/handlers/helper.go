package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/intraware/rodan/internal/utils/values"
)

type AuthService struct{}

func sendRequestWithRetry(method, fullURL, apiKey string, retries uint, delay time.Duration, timeout time.Duration) error {
	client := &http.Client{Timeout: timeout}
	for attempt := range retries {
		req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(nil))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", apiKey)
		resp, err := client.Do(req)
		if err != nil {
			if attempt < retries {
				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("failed to send request after %d attempts: %w", attempt, err)
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			if attempt < retries {
				time.Sleep(delay)
				continue
			}
			return fmt.Errorf("request failed with status %s after %d attempts", resp.Status, attempt)
		}
		return nil
	}
	return fmt.Errorf("request failed after %d retries", retries)
}

func buildAuthURL(baseURL, endpoint, route string) string {
	return fmt.Sprintf("%s%s/auth/%s", baseURL, endpoint, route)
}

func (a *AuthService) OpenLogin() error {
	cfg := values.GetConfig().App.Auth
	fullURL := buildAuthURL(cfg.URL, cfg.Endpoint, "open/login")
	return sendRequestWithRetry("POST", fullURL, cfg.HashedAPIKey, cfg.DefaultRetry, cfg.RetryDelay, cfg.Timeout)
}

func (a *AuthService) CloseLogin() error {
	cfg := values.GetConfig().App.Auth
	fullURL := buildAuthURL(cfg.URL, cfg.Endpoint, "close/login")
	return sendRequestWithRetry("POST", fullURL, cfg.HashedAPIKey, cfg.DefaultRetry, cfg.RetryDelay, cfg.Timeout)
}

func (a *AuthService) OpenSignup() error {
	cfg := values.GetConfig().App.Auth
	fullURL := buildAuthURL(cfg.URL, cfg.Endpoint, "open/signup")
	return sendRequestWithRetry("POST", fullURL, cfg.HashedAPIKey, cfg.DefaultRetry, cfg.RetryDelay, cfg.Timeout)
}

func (a *AuthService) CloseSignup() error {
	cfg := values.GetConfig().App.Auth
	fullURL := buildAuthURL(cfg.URL, cfg.Endpoint, "close/signup")
	return sendRequestWithRetry("POST", fullURL, cfg.HashedAPIKey, cfg.DefaultRetry, cfg.RetryDelay, cfg.Timeout)
}
