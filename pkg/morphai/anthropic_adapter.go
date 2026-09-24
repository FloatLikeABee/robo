package morphai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

const anthropicVersion = "2023-06-01"

type anthropicAdapter struct{}

type anthRequest struct {
	Model     string     `json:"model"`
	MaxTokens int        `json:"max_tokens"`
	System    string     `json:"system,omitempty"`
	Messages  []anthMsg  `json:"messages"`
	Tools     []anthTool `json:"tools,omitempty"`
	Stream    bool       `json:"stream,omitempty"`
}

type anthMsg struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type anthTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthToolUseBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type anthToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type anthImageBlock struct {
	Type   string       `json:"type"`
	Source anthImageSrc `json:"source"`
}

type anthImageSrc struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (anthropicAdapter) complete(ctx context.Context, c *Client, rc resolved, req outbound) (CompletionResponse, error) {
	payload, err := marshalAnthropic(rc, req, false)
	if err != nil {
		return CompletionResponse{}, err
	}
	hc := c.httpClient
	if req.Long {
		hc = c.httpClientLong
	}
	body, err := c.doJSON(ctx, hc, rc, anthropicMessagesURL(rc.BaseURL), payload)
	if err != nil {
		return CompletionResponse{}, err
	}
	resp, err := parseAnthropicCompletion(body)
	if err != nil {
		return CompletionResponse{}, withProvider(rc.Provider, err)
	}
	return resp, nil
}

func (anthropicAdapter) stream(ctx context.Context, c *Client, rc resolved, req outbound) (<-chan StreamEvent, error) {
	payload, err := marshalAnthropic(rc, req, true)
	if err != nil {
		return nil, err
	}
	hc := c.httpClient
	if req.Long {
		hc = c.httpClientLong
	}
	resp, err := c.doStream(ctx, hc, rc, anthropicMessagesURL(rc.BaseURL), payload)
	if err != nil {
		return nil, err
	}
	ch := make(chan StreamEvent, 16)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		streamAnthropic(ctx, rc, resp.Body, ch)
	}()
	return ch, nil
}

func marshalAnthropic(rc resolved, req outbound, stream bool) ([]byte, error) {
	var system string
	var msgs []anthMsg
	var err error
	if req.Vision {
		system, msgs, err = convertVision(req.Multi)
	} else {
		system, msgs, err = convertMessages(req.Messages)
	}
	if err != nil {
		return nil, err
	}
	body := anthRequest{
		Model:     req.UseModel,
		MaxTokens: anthMaxTokens(req),
		System:    system,
		Messages:  msgs,
		Stream:    stream,
	}
	if len(req.Tools) > 0 && !req.Vision {
		body.Tools = make([]anthTool, 0, len(req.Tools))
		for _, tool := range req.Tools {
			body.Tools = append(body.Tools, anthTool{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: rawObject(tool.Parameters),
			})
		}
	}
	return json.Marshal(body)
}

func anthMaxTokens(req outbound) int {
	if req.MaxTokens > 0 {
		return req.MaxTokens
	}
	if req.Long {
		return 8192
	}
	return 4096
}

func anthropicMessagesURL(base string) string {
	raw := strings.TrimSpace(base)
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		b := strings.TrimRight(raw, "/")
		lower := strings.ToLower(b)
		switch {
		case strings.HasSuffix(lower, "/messages"):
			return b
		case strings.HasSuffix(lower, "/v1"):
			return b + "/messages"
		default:
			return b + "/v1/messages"
		}
	}
	path := strings.TrimRight(u.Path, "/")
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, "/messages"):
		u.Path = path
	case strings.HasSuffix(lower, "/v1"):
		u.Path = path + "/messages"
	default:
		if path == "" {
			u.Path = "/v1/messages"
		} else {
			u.Path = path + "/v1/messages"
		}
	}
	return u.String()
}

func convertMessages(msgs []Message) (string, []anthMsg, error) {
	var sys []string
	var out []anthMsg
	var pending []any
	flush := func() {
		if len(pending) == 0 {
			return
		}
		out = appendAnth(out, "user", append([]any(nil), pending...))
		pending = nil
	}
	for _, m := range msgs {
		switch strings.ToLower(strings.TrimSpace(m.Role)) {
		case "system":
			if strings.TrimSpace(m.Content) != "" {
				sys = append(sys, m.Content)
			}
		case "tool":
			pending = append(pending, anthToolResultBlock{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   m.Content,
			})
		case "assistant":
			flush()
			blocks := assistantBlocks(m)
			if len(blocks) == 0 {
				continue
			}
			out = appendAnth(out, "assistant", blocks)
		default:
			flush()
			if strings.TrimSpace(m.Content) == "" {
				continue
			}
			out = appendAnth(out, "user", m.Content)
		}
	}
	flush()
	if len(out) == 0 {
		return "", nil, fmt.Errorf("no messages")
	}
	return strings.Join(sys, "\n\n"), out, nil
}

func assistantBlocks(m Message) []any {
	var blocks []any
	if m.Content != "" {
		blocks = append(blocks, anthTextBlock{Type: "text", Text: m.Content})
	}
	for _, tc := range m.ToolCalls {
		input := bytes.TrimSpace([]byte(tc.Arguments))
		if len(input) == 0 || !json.Valid(input) || input[0] != '{' {
			input = []byte("{}")
		}
		blocks = append(blocks, anthToolUseBlock{
			Type:  "tool_use",
			ID:    tc.ID,
			Name:  tc.Name,
			Input: json.RawMessage(input),
		})
	}
	return blocks
}

func convertVision(messages []MultiMessage) (string, []anthMsg, error) {
	var sys []string
	var out []anthMsg
	for _, m := range messages {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		if role == "" {
			role = "user"
		}
		if role == "system" {
			for _, p := range m.Content {
				if p.Type == ContentPartText && strings.TrimSpace(p.Text) != "" {
					sys = append(sys, p.Text)
				}
			}
			continue
		}
		blocks := anthBlocksFromMulti(m)
		if len(blocks) == 0 {
			continue
		}
		out = appendAnth(out, role, blocks)
	}
	if len(out) == 0 {
		return "", nil, fmt.Errorf("no messages")
	}
	return strings.Join(sys, "\n\n"), out, nil
}

func anthBlocksFromMulti(m MultiMessage) []any {
	var blocks []any
	for _, p := range m.Content {
		switch p.Type {
		case ContentPartText:
			if p.Text == "" {
				continue
			}
			blocks = append(blocks, anthTextBlock{Type: "text", Text: p.Text})
		case ContentPartImageURL:
			if p.ImageURL == nil || strings.TrimSpace(p.ImageURL.URL) == "" {
				continue
			}
			if media, data, ok := parseDataURL(p.ImageURL.URL); ok {
				blocks = append(blocks, anthImageBlock{
					Type: "image",
					Source: anthImageSrc{
						Type:      "base64",
						MediaType: media,
						Data:      data,
					},
				})
				continue
			}
			blocks = append(blocks, anthImageBlock{
				Type: "image",
				Source: anthImageSrc{
					Type: "url",
					URL:  p.ImageURL.URL,
				},
			})
		}
	}
	return blocks
}

func parseDataURL(u string) (mediaType, data string, ok bool) {
	if !strings.HasPrefix(u, "data:") {
		return "", "", false
	}
	rest := strings.TrimPrefix(u, "data:")
	comma := strings.Index(rest, ",")
	if comma < 0 {
		return "", "", false
	}
	meta := rest[:comma]
	if !strings.Contains(strings.ToLower(meta), "base64") {
		return "", "", false
	}
	mediaType = meta
	if i := strings.Index(mediaType, ";"); i >= 0 {
		mediaType = mediaType[:i]
	}
	if strings.TrimSpace(mediaType) == "" {
		mediaType = "image/jpeg"
	}
	return mediaType, rest[comma+1:], true
}

func appendAnth(msgs []anthMsg, role string, content any) []anthMsg {
	if role != "assistant" {
		role = "user"
	}
	content = normalizeAnthContent(content)
	if len(msgs) > 0 && msgs[len(msgs)-1].Role == role {
		msgs[len(msgs)-1].Content = mergeAnthContent(msgs[len(msgs)-1].Content, content)
		return msgs
	}
	return append(msgs, anthMsg{Role: role, Content: content})
}

func normalizeAnthContent(content any) any {
	blocks, ok := content.([]any)
	if !ok || len(blocks) != 1 {
		return content
	}
	if b, ok := blocks[0].(anthTextBlock); ok {
		return b.Text
	}
	return content
}

func mergeAnthContent(a, b any) any {
	return append(asAnthBlocks(a), asAnthBlocks(b)...)
}

func asAnthBlocks(v any) []any {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		if t == "" {
			return nil
		}
		return []any{anthTextBlock{Type: "text", Text: t}}
	case []any:
		return t
	default:
		return []any{t}
	}
}

type anthResponse struct {
	Content []struct {
		Type  string          `json:"type"`
		Text  string          `json:"text"`
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func parseAnthropicCompletion(body []byte) (CompletionResponse, error) {
	var parsed anthResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" && len(parsed.Content) == 0 {
		return CompletionResponse{}, &APIError{StatusCode: 200, Code: parsed.Error.Type, Message: parsed.Error.Message}
	}
	if len(parsed.Content) == 0 {
		return CompletionResponse{}, fmt.Errorf("no response from AI model")
	}
	return anthContentToResponse(parsed.Content, parsed.StopReason), nil
}

func anthContentToResponse(content []struct {
	Type  string          `json:"type"`
	Text  string          `json:"text"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}, stop string) CompletionResponse {
	var b strings.Builder
	var calls []ToolCall
	for _, block := range content {
		switch block.Type {
		case "text":
			b.WriteString(block.Text)
		case "tool_use":
			calls = append(calls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: string(bytes.TrimSpace(block.Input)),
			})
		}
	}
	return CompletionResponse{
		Content:      b.String(),
		ToolCalls:    calls,
		FinishReason: mapAnthropicStop(stop),
	}
}

func mapAnthropicStop(reason string) string {
	switch reason {
	case "end_turn", "stop_sequence":
		return "stop"
	case "tool_use":
		return "tool_calls"
	case "max_tokens":
		return "length"
	default:
		return reason
	}
}

type anthStreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
		StopReason  string `json:"stop_reason"`
	} `json:"delta"`
	ContentBlock struct {
		Type string `json:"type"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Text string `json:"text"`
	} `json:"content_block"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

type anthToolAcc struct {
	id     string
	name   string
	args   strings.Builder
	isTool bool
}

func streamAnthropic(ctx context.Context, rc resolved, r io.Reader, ch chan<- StreamEvent) {
	acc := map[int]*anthToolAcc{}
	var finish string
	err := readSSE(r, func(ev sseEvent) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data := strings.TrimSpace(ev.Data)
		if data == "" || data == "[DONE]" {
			if data == "[DONE]" {
				return errSSEDone
			}
			return nil
		}
		var parsed anthStreamEvent
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			return fmt.Errorf("unmarshal stream: %w", err)
		}
		typ := parsed.Type
		if typ == "" {
			typ = ev.Event
		}
		if parsed.Error != nil && parsed.Error.Message != "" {
			return &APIError{Provider: rc.Provider, StatusCode: 200, Code: parsed.Error.Type, Message: parsed.Error.Message}
		}
		switch typ {
		case "content_block_start":
			if parsed.ContentBlock.Type == "tool_use" {
				acc[parsed.Index] = &anthToolAcc{id: parsed.ContentBlock.ID, name: parsed.ContentBlock.Name, isTool: true}
			} else if parsed.ContentBlock.Text != "" {
				if !sendEvent(ctx, ch, StreamEvent{Delta: parsed.ContentBlock.Text}) {
					return ctx.Err()
				}
			}
		case "content_block_delta":
			switch parsed.Delta.Type {
			case "text_delta":
				if parsed.Delta.Text != "" {
					if !sendEvent(ctx, ch, StreamEvent{Delta: parsed.Delta.Text}) {
						return ctx.Err()
					}
				}
			case "input_json_delta":
				tool := acc[parsed.Index]
				if tool == nil {
					tool = &anthToolAcc{isTool: true}
					acc[parsed.Index] = tool
				}
				tool.args.WriteString(parsed.Delta.PartialJSON)
			}
		case "message_delta":
			if parsed.Delta.StopReason != "" {
				finish = mapAnthropicStop(parsed.Delta.StopReason)
			}
		case "message_stop":
			return errSSEDone
		case "error":
			msg := "stream error"
			code := "error"
			if parsed.Error != nil {
				if parsed.Error.Message != "" {
					msg = parsed.Error.Message
				}
				if parsed.Error.Type != "" {
					code = parsed.Error.Type
				}
			}
			return &APIError{Provider: rc.Provider, StatusCode: 200, Code: code, Message: msg}
		}
		return nil
	})
	if err != nil && !errors.Is(err, errSSEDone) {
		if ctx.Err() != nil {
			sendTerminal(ctx, ch, StreamEvent{Err: ctx.Err()})
			return
		}
		sendTerminal(ctx, ch, StreamEvent{Err: err})
		return
	}
	if ctx.Err() != nil {
		sendTerminal(ctx, ch, StreamEvent{Err: ctx.Err()})
		return
	}
	var calls []ToolCall
	if len(acc) > 0 {
		idxs := make([]int, 0, len(acc))
		for idx := range acc {
			idxs = append(idxs, idx)
		}
		sort.Ints(idxs)
		for _, i := range idxs {
			tool := acc[i]
			if tool == nil || !tool.isTool {
				continue
			}
			calls = append(calls, ToolCall{ID: tool.id, Name: tool.name, Arguments: normalizeToolArgs(tool.args.String())})
		}
	}
	sendTerminal(ctx, ch, StreamEvent{Done: true, FinishReason: finish, ToolCalls: calls})
}
