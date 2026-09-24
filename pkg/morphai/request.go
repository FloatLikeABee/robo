package morphai

import (
	"encoding/json"
	"strings"
)

// Tool is a function the model may call. Parameters is a JSON Schema object.
// Adapters translate this into OpenAI tools or Anthropic tool definitions.
type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolCall is one function invocation requested by the model.
// Arguments is a JSON object encoded as text.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// CompletionRequest is one chat call. Empty provider, model, key, and base URL
// fields inherit from the client. Setting Provider to a different id does not
// inherit the client's key or base URL.
type CompletionRequest struct {
	Provider ProviderID
	Model    string
	APIKey   string
	BaseURL  string

	Messages []Message
	Tools    []Tool
	// JSONMode asks for a JSON object response. Providers without a JSON
	// response format return ErrCapabilityUnsupported.
	JSONMode  bool
	MaxTokens int
}

// CompletionResponse is the model's reply. Content and ToolCalls can both be set.
type CompletionResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
}

// StreamEvent is one piece of a streaming reply.
// Delta is a text fragment. ToolCalls and FinishReason are set on the event
// that has Done set. Err is terminal; the channel is then closed.
type StreamEvent struct {
	Delta        string
	ToolCalls    []ToolCall
	FinishReason string
	Done         bool
	Err          error
}

// CollectStream reads ch until it closes and joins the text deltas.
func CollectStream(ch <-chan StreamEvent) (CompletionResponse, error) {
	var b strings.Builder
	var out CompletionResponse
	if ch == nil {
		return out, errNilStream
	}
	for ev := range ch {
		if ev.Err != nil {
			return CompletionResponse{}, ev.Err
		}
		b.WriteString(ev.Delta)
		if len(ev.ToolCalls) > 0 {
			out.ToolCalls = ev.ToolCalls
		}
		if ev.FinishReason != "" {
			out.FinishReason = ev.FinishReason
		}
	}
	out.Content = b.String()
	return out, nil
}

type errString string

func (e errString) Error() string { return string(e) }

const errNilStream = errString("nil stream")
