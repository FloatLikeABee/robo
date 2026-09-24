package morphai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type toolAcc struct {
	id   string
	name string
	args strings.Builder
}

func streamOpenAI(ctx context.Context, rc resolved, r io.Reader, ch chan<- StreamEvent) {
	var acc []toolAcc
	var finish string
	err := readSSE(r, func(ev sseEvent) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		data := strings.TrimSpace(ev.Data)
		if data == "" {
			return nil
		}
		if data == "[DONE]" {
			return errSSEDone
		}
		var chunk oaiStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return fmt.Errorf("unmarshal stream: %w", err)
		}
		if chunk.Error != nil && chunk.Error.Message != "" {
			code := chunk.Error.Code
			if code == "" {
				code = chunk.Error.Type
			}
			return &APIError{Provider: rc.Provider, StatusCode: 200, Code: code, Message: chunk.Error.Message}
		}
		if len(chunk.Choices) == 0 {
			return nil
		}
		choice := chunk.Choices[0]
		if choice.FinishReason != "" {
			finish = choice.FinishReason
		}
		if delta := decodeContent(choice.Delta.Content); delta != "" {
			if !sendEvent(ctx, ch, StreamEvent{Delta: delta}) {
				return ctx.Err()
			}
		}
		for _, tc := range choice.Delta.ToolCalls {
			if tc.Index < 0 || tc.Index > 64 {
				continue
			}
			for len(acc) <= tc.Index {
				acc = append(acc, toolAcc{})
			}
			if tc.ID != "" {
				acc[tc.Index].id = tc.ID
			}
			if tc.Function.Name != "" {
				acc[tc.Index].name += tc.Function.Name
			}
			if args := decodeArgs(tc.Function.Arguments); args != "" {
				acc[tc.Index].args.WriteString(args)
			}
		}
		return nil
	})
	if err != nil && !errors.Is(err, errSSEDone) {
		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			sendEvent(context.Background(), ch, StreamEvent{Err: ctx.Err()})
			return
		}
		sendEvent(ctx, ch, StreamEvent{Err: err})
		return
	}
	if ctx.Err() != nil {
		sendEvent(context.Background(), ch, StreamEvent{Err: ctx.Err()})
		return
	}
	sendEvent(ctx, ch, StreamEvent{Done: true, FinishReason: finish, ToolCalls: accToolCalls(acc)})
}

func accToolCalls(acc []toolAcc) []ToolCall {
	if len(acc) == 0 {
		return nil
	}
	out := make([]ToolCall, 0, len(acc))
	for _, tc := range acc {
		if tc.id == "" && tc.name == "" && tc.args.Len() == 0 {
			continue
		}
		out = append(out, ToolCall{ID: tc.id, Name: tc.name, Arguments: tc.args.String()})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sendEvent(ctx context.Context, ch chan<- StreamEvent, ev StreamEvent) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- ev:
		return true
	}
}

// errSSEDone stops the SSE reader at the OpenAI [DONE] marker.
var errSSEDone = errString("sse done")
