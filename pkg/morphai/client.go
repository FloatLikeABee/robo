package morphai

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client calls the configured AI provider. An empty Config.Provider keeps
// today's DashScope / MORPH_AI_* behavior.
type Client struct {
	cfg                Config
	httpClient         *http.Client
	httpClientLong     *http.Client
	lastRequestTime    time.Time
	requestMutex       sync.Mutex
	minRequestInterval time.Duration
	maxRetries         int
	retryBase          time.Duration
}

// NewClient builds a client from explicit config.
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
		maxRetries:         3,
		retryBase:          2 * time.Second,
	}
}

// NewClientFromEnv loads config from the environment and returns a client.
func NewClientFromEnv() *Client {
	return NewClient(LoadFromEnv())
}

// Configured reports whether a call can be made with this client's config.
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
	if normalizeProviderID(c.cfg.Provider) == "" {
		return c.cfg.VisionModelOrDefault()
	}
	rc, _ := resolve(c.cfg)
	if rc.VisionModel != "" {
		return rc.VisionModel
	}
	return c.cfg.VisionModelOrDefault()
}

// ResolvedConfig returns the provider settings a call with this client would use.
func (c *Client) ResolvedConfig() (ResolvedConfig, error) {
	if c == nil {
		return ResolvedConfig{}, fmt.Errorf("morphai client is nil")
	}
	return ResolveConfig(c.cfg)
}

type outbound struct {
	CompletionRequest
	Multi    []MultiMessage
	Long     bool
	Vision   bool
	UseModel string
}

type chatAdapter interface {
	complete(ctx context.Context, c *Client, rc resolved, req outbound) (CompletionResponse, error)
	stream(ctx context.Context, c *Client, rc resolved, req outbound) (<-chan StreamEvent, error)
}

func adapterFor(kind adapterKind) chatAdapter {
	switch kind {
	case adapterAnthropic:
		return anthropicAdapter{}
	case adapterDashScopeNative:
		return dashScopeNativeAdapter{}
	default:
		return openAIAdapter{}
	}
}

// ChatCompletion sends messages to the configured model and returns the reply text.
func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	resp, err := c.complete(ctx, outbound{CompletionRequest: CompletionRequest{Messages: messages}})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// ChatCompletionLong uses a longer HTTP timeout for heavy generation tasks.
func (c *Client) ChatCompletionLong(ctx context.Context, messages []Message) (string, error) {
	resp, err := c.complete(ctx, outbound{CompletionRequest: CompletionRequest{Messages: messages}, Long: true})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Complete sends a chat request, including tools and JSON mode when set.
func (c *Client) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return c.complete(ctx, outbound{CompletionRequest: req})
}

// CompleteLong is Complete with the longer HTTP timeout.
func (c *Client) CompleteLong(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	return c.complete(ctx, outbound{CompletionRequest: req, Long: true})
}

// ChatCompletionVision sends multimodal messages (text plus images) to a
// vision-capable model and returns the reply text.
//
// A client configured for the native DashScope text-generation endpoint returns
// an error rather than sending a payload that endpoint cannot parse.
//
// Pass an empty model to use the configured vision model.
func (c *Client) ChatCompletionVision(ctx context.Context, messages []MultiMessage, model string) (string, error) {
	resp, err := c.complete(ctx, outbound{
		CompletionRequest: CompletionRequest{Model: model},
		Multi:             messages,
		Vision:            true,
		Long:              true,
	})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Stream sends a streaming chat request. The channel is closed after a
// terminal event. Setup failures return a nil channel and an error.
func (c *Client) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamEvent, error) {
	if c == nil {
		return nil, fmt.Errorf("morphai client is nil")
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("no messages")
	}
	rc, err := resolve(c.cfg.applyCall(req))
	if err != nil {
		return nil, err
	}
	if err := rc.reject(capStream); err != nil {
		return nil, err
	}
	if len(req.Tools) > 0 {
		if err := rc.reject(capTools); err != nil {
			return nil, err
		}
	}
	if req.JSONMode {
		if err := rc.reject(capJSON); err != nil {
			return nil, err
		}
	}
	out := outbound{CompletionRequest: req, UseModel: rc.Model}
	return adapterFor(rc.Kind).stream(ctx, c, rc, out)
}

func (c *Client) complete(ctx context.Context, req outbound) (CompletionResponse, error) {
	if c == nil {
		return CompletionResponse{}, fmt.Errorf("morphai client is nil")
	}
	if req.Vision {
		if len(req.Multi) == 0 {
			return CompletionResponse{}, fmt.Errorf("no messages")
		}
	} else if len(req.Messages) == 0 {
		return CompletionResponse{}, fmt.Errorf("no messages")
	}

	call := req.CompletionRequest
	visionModel := ""
	if req.Vision {
		// The vision model argument must not replace the chat model.
		visionModel = strings.TrimSpace(call.Model)
		call.Model = ""
	}
	rc, err := resolve(c.cfg.applyCall(call))
	if err != nil {
		return CompletionResponse{}, err
	}
	if req.Vision {
		if err := rc.reject(capVision); err != nil {
			return CompletionResponse{}, err
		}
		if visionModel != "" {
			req.UseModel = visionModel
		} else {
			req.UseModel = rc.VisionModel
		}
	} else {
		if err := rc.reject(capChat); err != nil {
			return CompletionResponse{}, err
		}
		if len(req.Tools) > 0 {
			if err := rc.reject(capTools); err != nil {
				return CompletionResponse{}, err
			}
		}
		if req.JSONMode {
			if err := rc.reject(capJSON); err != nil {
				return CompletionResponse{}, err
			}
		}
		req.UseModel = rc.Model
	}
	return adapterFor(rc.Kind).complete(ctx, c, rc, req)
}

// DataURL renders raw bytes as a base64 data URL for a multimodal image part.
func DataURL(mimeType string, raw []byte) string {
	m := strings.TrimSpace(mimeType)
	if !strings.HasPrefix(m, "image/") {
		m = "image/jpeg"
	}
	return "data:" + m + ";base64," + base64.StdEncoding.EncodeToString(raw)
}
