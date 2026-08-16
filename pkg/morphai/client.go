package morphai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client calls Alibaba DashScope text-generation (same stack as Morph Data).
type Client struct {
	cfg                  Config
	httpClient           *http.Client
	httpClientLong       *http.Client
	lastRequestTime      time.Time
	requestMutex         sync.Mutex
	minRequestInterval   time.Duration
}

// NewClient builds a DashScope client from explicit config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		httpClientLong: &http.Client{
			Timeout: 300 * time.Second,
		},
		minRequestInterval: 200 * time.Millisecond,
	}
}

// NewClientFromEnv loads config from the environment and returns a client.
func NewClientFromEnv() *Client {
	return NewClient(LoadFromEnv())
}

// Configured reports whether an API key is set.
func (c *Client) Configured() bool {
	if c == nil {
		return false
	}
	return c.cfg.Configured()
}

// VisionSupported reports whether this client can send image content.
func (c *Client) VisionSupported() bool {
	if c == nil {
		return false
	}
	return c.cfg.VisionSupported()
}

// VisionModel returns the multimodal model this client uses for image requests.
func (c *Client) VisionModel() string {
	if c == nil {
		return ""
	}
	return c.cfg.VisionModelOrDefault()
}

type dashScopeRequest struct {
	Model string `json:"model"`
	Input struct {
		Messages []Message `json:"messages"`
	} `json:"input"`
}

type dashScopeResponse struct {
	Output struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (c *Client) rateLimit() {
	c.requestMutex.Lock()
	defer c.requestMutex.Unlock()
	now := time.Now()
	if wait := c.minRequestInterval - now.Sub(c.lastRequestTime); wait > 0 {
		time.Sleep(wait)
	}
	c.lastRequestTime = time.Now()
}

// ChatCompletion sends messages to the configured model without caching.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	return c.chatCompletionWithClient(ctx, messages, c.httpClient)
}

// ChatCompletionLong uses a longer HTTP timeout for heavy generation tasks.
func (c *Client) ChatCompletionLong(ctx context.Context, messages []Message) (string, error) {
	return c.chatCompletionWithClient(ctx, messages, c.httpClientLong)
}

func (c *Client) chatCompletionWithClient(ctx context.Context, messages []Message, client *http.Client) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages")
	}
	if !c.cfg.Configured() {
		return "", fmt.Errorf("MORPH_AI_API_KEY is not configured")
	}
	if !c.cfg.UseNativeAPI {
		return c.chatOpenAICompat(ctx, messages, client)
	}

	reqBody := dashScopeRequest{Model: c.cfg.Model}
	reqBody.Input.Messages = messages
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}
		c.rateLimit()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.APIURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue
			}
			return "", fmt.Errorf("send request: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if attempt < maxRetries {
				continue
			}
			return "", fmt.Errorf("read response: %w", readErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			var errResp struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
				return "", fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Code, errResp.Message)
			}
			return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		var parsed dashScopeResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return "", fmt.Errorf("unmarshal response: %w", err)
		}
		if parsed.Code != "" && parsed.Code != "Success" {
			return "", fmt.Errorf("API error: %s - %s", parsed.Code, parsed.Message)
		}
		if len(parsed.Output.Choices) == 0 {
			return "", fmt.Errorf("no response from AI model")
		}
		return parsed.Output.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("max retries exceeded")
}

type openAIChatRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	EnableThinking *bool     `json:"enable_thinking,omitempty"`
}

type openAIVisionRequest struct {
	Model    string         `json:"model"`
	Messages []MultiMessage `json:"messages"`
}

// ChatCompletionVision sends multimodal messages (text plus images) to a
// vision-capable model and returns the reply text.
//
// Only the OpenAI-compatible endpoint accepts a content array, so a client
// configured for the native DashScope text-generation endpoint returns an error
// rather than sending a payload the endpoint cannot parse.
//
// Pass an empty model to use the configured vision model.
func (c *Client) ChatCompletionVision(ctx context.Context, messages []MultiMessage, model string) (string, error) {
	if len(messages) == 0 {
		return "", fmt.Errorf("no messages")
	}
	if !c.cfg.Configured() {
		return "", fmt.Errorf("MORPH_AI_API_KEY is not configured")
	}
	if c.cfg.UseNativeAPI {
		return "", fmt.Errorf(
			"vision requests need an OpenAI-compatible endpoint; MORPH_AI_API_URL is set to the native DashScope text-generation endpoint. " +
				"Unset MORPH_AI_API_URL (or point it at a /v1 compatible base URL) and set MORPH_AI_BASE_URL to enable image reading")
	}

	if strings.TrimSpace(model) == "" {
		model = c.cfg.VisionModelOrDefault()
	}
	reqBody := openAIVisionRequest{Model: model, Messages: messages}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	return c.postOpenAIChat(ctx, jsonData, c.httpClientLong)
}

// DataURL renders raw bytes as a base64 data URL for a multimodal image part.
func DataURL(mimeType string, raw []byte) string {
	m := strings.TrimSpace(mimeType)
	if !strings.HasPrefix(m, "image/") {
		m = "image/jpeg"
	}
	return "data:" + m + ";base64," + base64.StdEncoding.EncodeToString(raw)
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (c *Client) chatOpenAICompat(ctx context.Context, messages []Message, client *http.Client) (string, error) {
	// Qwen3.x reasoning models otherwise spend minutes on chain-of-thought before content.
	thinkingOff := false
	reqBody := openAIChatRequest{Model: c.cfg.Model, Messages: messages, EnableThinking: &thinkingOff}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	return c.postOpenAIChat(ctx, jsonData, client)
}

func (c *Client) postOpenAIChat(ctx context.Context, jsonData []byte, client *http.Client) (string, error) {
	endpoint := c.cfg.BaseURL + "/chat/completions"

	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			time.Sleep(delay)
		}
		c.rateLimit()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				continue
			}
			return "", fmt.Errorf("send request: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			if attempt < maxRetries {
				continue
			}
			return "", fmt.Errorf("read response: %w", readErr)
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			var parsed openAIChatResponse
			if json.Unmarshal(body, &parsed) == nil && parsed.Error != nil && parsed.Error.Message != "" {
				code := parsed.Error.Code
				if code == "" {
					code = parsed.Error.Type
				}
				return "", fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, code, parsed.Error.Message)
			}
			var errResp struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if json.Unmarshal(body, &errResp) == nil && errResp.Message != "" {
				return "", fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, errResp.Code, errResp.Message)
			}
			return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		var parsed openAIChatResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return "", fmt.Errorf("unmarshal response: %w", err)
		}
		if len(parsed.Choices) == 0 {
			return "", fmt.Errorf("no response from AI model")
		}
		return parsed.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("max retries exceeded")
}
