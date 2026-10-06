package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// EventLogResult is the Morph events API response, trimmed so a token cannot hide in it.
type EventLogResult struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

func eventAPIBase() string {
	base := strings.TrimSpace(os.Getenv("MORPH_API_BASE_URL"))
	if base == "" {
		base = "http://127.0.0.1:9090"
	}
	return strings.TrimRight(base, "/")
}

func eventAPIToken() string {
	return strings.TrimSpace(os.Getenv(TokenEnv))
}

// CreateEventLog posts one event through the Morph events API.
func CreateEventLog(ctx context.Context, title, when, detail, reporter string) (EventLogResult, error) {
	title = strings.TrimSpace(title)
	when = strings.TrimSpace(when)
	if title == "" || when == "" {
		return EventLogResult{}, errors.New("title and time are required")
	}
	return callEventAPI(ctx, http.MethodPost, "/api/sheetx/events-info", map[string]string{
		"title":    title,
		"time":     when,
		"detail":   strings.TrimSpace(detail),
		"reporter": strings.TrimSpace(reporter),
	})
}

// ListEventLogs reads a page of event logs.
func ListEventLogs(ctx context.Context, limit int) (EventLogResult, error) {
	n, err := appliedLimit(limit)
	if err != nil {
		return EventLogResult{}, err
	}
	return callEventAPI(ctx, http.MethodGet, "/api/sheetx/events-info?page=1&limit="+strconv.Itoa(n), nil)
}

// GetEventLog reads one event log.
func GetEventLog(ctx context.Context, id string) (EventLogResult, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return EventLogResult{}, errors.New("invalid id")
	}
	return callEventAPI(ctx, http.MethodGet, "/api/sheetx/events-info/"+id, nil)
}

func callEventAPI(ctx context.Context, method, path string, body any) (EventLogResult, error) {
	token := eventAPIToken()
	if token == "" {
		return EventLogResult{}, errors.New("event log bridge is not configured")
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return EventLogResult{}, errors.New("could not encode the event log")
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, eventAPIBase()+path, reader)
	if err != nil {
		return EventLogResult{}, errors.New("could not reach event logs")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return EventLogResult{}, errors.New("could not reach event logs")
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 4000))
	text := string(raw)
	if strings.Contains(text, token) {
		text = strings.ReplaceAll(text, token, "[redacted]")
	}
	out := EventLogResult{Status: res.StatusCode, Body: text}
	if res.StatusCode >= 300 {
		return out, errors.New("event log request failed")
	}
	return out, nil
}
