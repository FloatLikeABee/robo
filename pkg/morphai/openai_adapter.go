package morphai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type openAIAdapter struct{}

type oaiMessage struct {
	Role       string          `json:"role"`
	Content    *string         `json:"content,omitempty"`
	Name       string          `json:"name,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	ToolCalls  []oaiToolCallIn `json:"tool_calls,omitempty"`
}

type oaiToolCallIn struct {
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Function oaiFunctionCall `json:"function"`
}

type oaiFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type oaiToolDef struct {
	Type     string      `json:"type"`
	Function oaiFunction `json:"function"`
}

type oaiFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type oaiResponseFormat struct {
	Type string `json:"type"`
}

type oaiChatRequest struct {
	Model          string             `json:"model"`
	Messages       []oaiMessage       `json:"messages"`
	Tools          []oaiToolDef       `json:"tools,omitempty"`
	ResponseFormat *oaiResponseFormat `json:"response_format,omitempty"`
	Stream         bool               `json:"stream,omitempty"`
	MaxTokens      int                `json:"max_tokens,omitempty"`
	EnableThinking *bool              `json:"enable_thinking,omitempty"`
}

type oaiVisionRequest struct {
	Model    string         `json:"model"`
	Messages []MultiMessage `json:"messages"`
}

func (openAIAdapter) complete(ctx context.Context, c *Client, rc resolved, req outbound) (CompletionResponse, error) {
	payload, err := marshalOpenAI(rc, req, false)
	if err != nil {
		return CompletionResponse{}, err
	}
	hc := c.httpClient
	if req.Long {
		hc = c.httpClientLong
	}
	body, err := c.doJSON(ctx, hc, rc, openAIChatURL(rc.BaseURL), payload)
	if err != nil {
		return CompletionResponse{}, err
	}
	resp, err := parseOpenAICompletion(body)
	if err != nil {
		return CompletionResponse{}, withProvider(rc.Provider, err)
	}
	return resp, nil
}

func (openAIAdapter) stream(ctx context.Context, c *Client, rc resolved, req outbound) (<-chan StreamEvent, error) {
	payload, err := marshalOpenAI(rc, req, true)
	if err != nil {
		return nil, err
	}
	hc := c.httpClient
	if req.Long {
		hc = c.httpClientLong
	}
	resp, err := c.doStream(ctx, hc, rc, openAIChatURL(rc.BaseURL), payload)
	if err != nil {
		return nil, err
	}
	ch := make(chan StreamEvent, 16)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		streamOpenAI(ctx, rc, resp.Body, ch)
	}()
	return ch, nil
}

func marshalOpenAI(rc resolved, req outbound, stream bool) ([]byte, error) {
	if req.Vision {
		// Vision requests historically omit enable_thinking.
		return json.Marshal(oaiVisionRequest{Model: req.UseModel, Messages: req.Multi})
	}
	msgs := make([]oaiMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, toOAIMessage(m))
	}
	body := oaiChatRequest{
		Model:     req.UseModel,
		Messages:  msgs,
		Stream:    stream,
		MaxTokens: req.MaxTokens,
	}
	if len(req.Tools) > 0 {
		body.Tools = make([]oaiToolDef, 0, len(req.Tools))
		for _, tool := range req.Tools {
			body.Tools = append(body.Tools, oaiToolDef{
				Type: "function",
				Function: oaiFunction{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  rawObject(tool.Parameters),
				},
			})
		}
	}
	if req.JSONMode {
		body.ResponseFormat = &oaiResponseFormat{Type: "json_object"}
	}
	if rc.EnableThinkingOff {
		off := false
		body.EnableThinking = &off
	}
	return json.Marshal(body)
}

func toOAIMessage(m Message) oaiMessage {
	out := oaiMessage{
		Role:       m.Role,
		Name:       m.Name,
		ToolCallID: m.ToolCallID,
	}
	if m.Content != "" || len(m.ToolCalls) == 0 {
		content := m.Content
		out.Content = &content
	}
	for _, tc := range m.ToolCalls {
		args := tc.Arguments
		if strings.TrimSpace(args) == "" {
			args = "{}"
		}
		out.ToolCalls = append(out.ToolCalls, oaiToolCallIn{
			ID:   tc.ID,
			Type: "function",
			Function: oaiFunctionCall{
				Name:      tc.Name,
				Arguments: args,
			},
		})
	}
	return out
}

func rawObject(raw json.RawMessage) json.RawMessage {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage(`{"type":"object","properties":{}}`)
	}
	return raw
}

type oaiCompletion struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content   json.RawMessage `json:"content"`
			ToolCalls []struct {
				ID       string `json:"id"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func parseOpenAICompletion(body []byte) (CompletionResponse, error) {
	var parsed oaiCompletion
	if err := json.Unmarshal(body, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" && len(parsed.Choices) == 0 {
		code := parsed.Error.Code
		if code == "" {
			code = parsed.Error.Type
		}
		return CompletionResponse{}, &APIError{StatusCode: 200, Code: code, Message: parsed.Error.Message}
	}
	if len(parsed.Choices) == 0 {
		return CompletionResponse{}, fmt.Errorf("no response from AI model")
	}
	choice := parsed.Choices[0]
	resp := CompletionResponse{
		Content:      decodeContent(choice.Message.Content),
		FinishReason: choice.FinishReason,
	}
	for _, tc := range choice.Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: decodeArgs(tc.Function.Arguments),
		})
	}
	return resp, nil
}

func decodeContent(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Type == "" || p.Type == "text" {
				b.WriteString(p.Text)
			}
		}
		return b.String()
	}
	return ""
}

func decodeArgs(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return string(raw)
}

type oaiStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content   json.RawMessage `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}
