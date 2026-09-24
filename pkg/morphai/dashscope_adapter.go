package morphai

import (
	"context"
	"encoding/json"
	"fmt"
)

// nativeVisionDetail is the historical ChatCompletionVision error for the
// native DashScope text-generation endpoint. Callers match this text.
const nativeVisionDetail = "vision requests need an OpenAI-compatible endpoint; MORPH_AI_API_URL is set to the native DashScope text-generation endpoint. " +
	"Unset MORPH_AI_API_URL (or point it at a /v1 compatible base URL) and set MORPH_AI_BASE_URL to enable image reading"

type dashScopeNativeAdapter struct{}

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

func (dashScopeNativeAdapter) complete(ctx context.Context, c *Client, rc resolved, req outbound) (CompletionResponse, error) {
	if req.Vision {
		return CompletionResponse{}, rc.reject(capVision)
	}
	body := dashScopeRequest{Model: req.UseModel}
	body.Input.Messages = req.Messages
	payload, err := json.Marshal(body)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshal request: %w", err)
	}
	endpoint := rc.APIURL
	if endpoint == "" {
		endpoint = DefaultAPIURL
	}
	hc := c.httpClient
	if req.Long {
		hc = c.httpClientLong
	}
	raw, err := c.doJSON(ctx, hc, rc, endpoint, payload)
	if err != nil {
		return CompletionResponse{}, err
	}
	return parseDashScope(raw)
}

func (dashScopeNativeAdapter) stream(context.Context, *Client, resolved, outbound) (<-chan StreamEvent, error) {
	return nil, &CapabilityError{Provider: ProviderDashScope, Capability: capStream}
}

func parseDashScope(body []byte) (CompletionResponse, error) {
	var parsed dashScopeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Code != "" && parsed.Code != "Success" {
		return CompletionResponse{}, fmt.Errorf("API error: %s - %s", parsed.Code, parsed.Message)
	}
	if len(parsed.Output.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("no response from AI model")
	}
	return CompletionResponse{Content: parsed.Output.Choices[0].Message.Content, FinishReason: "stop"}, nil
}
