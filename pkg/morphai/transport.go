package morphai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (c *Client) rateLimit() {
	c.requestMutex.Lock()
	defer c.requestMutex.Unlock()
	now := time.Now()
	if wait := c.minRequestInterval - now.Sub(c.lastRequestTime); wait > 0 {
		time.Sleep(wait)
	}
	c.lastRequestTime = time.Now()
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func authHeader(rc resolved) http.Header {
	h := make(http.Header)
	h.Set("Content-Type", "application/json")
	switch rc.Kind {
	case adapterAnthropic:
		if rc.APIKey != "" {
			h.Set("x-api-key", rc.APIKey)
		}
		h.Set("anthropic-version", anthropicVersion)
	default:
		if rc.APIKey != "" {
			h.Set("Authorization", "Bearer "+rc.APIKey)
		}
	}
	return h
}

func (c *Client) doJSON(ctx context.Context, hc *http.Client, rc resolved, endpoint string, payload []byte) ([]byte, error) {
	maxRetries := c.maxRetries
	baseDelay := c.retryBase
	header := authHeader(rc)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if err := sleepCtx(ctx, delay); err != nil {
				return nil, err
			}
		}
		c.rateLimit()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		for k, vals := range header {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}

		resp, err := hc.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue
			}
			return nil, fmt.Errorf("send request: %w", err)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if attempt < maxRetries {
				continue
			}
			return nil, fmt.Errorf("read response: %w", readErr)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, withProvider(rc.Provider, parseProviderError(resp.StatusCode, body))
		}
		return body, nil
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func (c *Client) doStream(ctx context.Context, hc *http.Client, rc resolved, endpoint string, payload []byte) (*http.Response, error) {
	maxRetries := c.maxRetries
	baseDelay := c.retryBase
	header := authHeader(rc)
	header.Set("Accept", "text/event-stream")

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if err := sleepCtx(ctx, delay); err != nil {
				return nil, err
			}
		}
		c.rateLimit()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		for k, vals := range header {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}

		resp, err := hc.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue
			}
			return nil, fmt.Errorf("send request: %w", err)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			resp.Body.Close()
			continue
		}
		if resp.StatusCode != http.StatusOK {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				return nil, fmt.Errorf("read response: %w", readErr)
			}
			return nil, withProvider(rc.Provider, parseProviderError(resp.StatusCode, body))
		}
		return resp, nil
	}
	return nil, fmt.Errorf("max retries exceeded")
}

func withProvider(provider ProviderID, err error) error {
	var api *APIError
	if errors.As(err, &api) {
		api.Provider = provider
	}
	return err
}

func parseProviderError(status int, body []byte) error {
	var wrapped struct {
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &wrapped) == nil && wrapped.Error != nil && wrapped.Error.Message != "" {
		code := wrapped.Error.Code
		if code == "" {
			code = wrapped.Error.Type
		}
		return &APIError{StatusCode: status, Code: code, Message: wrapped.Error.Message}
	}
	var flat struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &flat) == nil && flat.Message != "" {
		return &APIError{StatusCode: status, Code: flat.Code, Message: flat.Message}
	}
	return fmt.Errorf("API returned status %d: %s", status, clipBody(body))
}

func clipBody(b []byte) string {
	const max = 8192
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "…"
}

func openAIChatURL(base string) string {
	b := strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(strings.ToLower(b), "/chat/completions") {
		return b
	}
	return b + "/chat/completions"
}
