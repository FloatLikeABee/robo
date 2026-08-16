package ai

import (
	"fmt"
	"strings"
)

// normalizeChatMessages merges leading system messages into one block.
// SiliconFlow and some OpenAI-compatible APIs reject multiple system roles (error 20015).
func normalizeChatMessages(msgs []ChatMessage) []ChatMessage {
	if len(msgs) == 0 {
		return msgs
	}
	var systemParts []string
	var rest []ChatMessage
	pastSystem := false
	for _, m := range msgs {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		content := strings.TrimSpace(m.Content)
		if !pastSystem && role == "system" {
			if content != "" {
				systemParts = append(systemParts, content)
			}
			continue
		}
		pastSystem = true
		if role == "system" {
			if content != "" {
				rest = append(rest, ChatMessage{
					Role:    "user",
					Content: "[Context]\n" + content,
				})
			}
			continue
		}
		if role == "user" || role == "assistant" {
			rest = append(rest, m)
		}
	}
	out := make([]ChatMessage, 0, 1+len(rest))
	if len(systemParts) > 0 {
		out = append(out, ChatMessage{Role: "system", Content: strings.Join(systemParts, "\n\n")})
	}
	return append(out, rest...)
}

// normalizeRawMessages is the map-based variant used by help-you-learn / multimodal calls.
func normalizeRawMessages(messages []map[string]interface{}) []map[string]interface{} {
	if len(messages) == 0 {
		return messages
	}
	var systemParts []string
	var rest []map[string]interface{}
	pastSystem := false
	for _, m := range messages {
		role, _ := m["role"].(string)
		role = strings.ToLower(strings.TrimSpace(role))
		if !pastSystem && role == "system" {
			if s := rawMessageText(m["content"]); s != "" {
				systemParts = append(systemParts, s)
			}
			continue
		}
		pastSystem = true
		if role == "system" {
			if s := rawMessageText(m["content"]); s != "" {
				rest = append(rest, map[string]interface{}{
					"role":    "user",
					"content": "[Context]\n" + s,
				})
			}
			continue
		}
		rest = append(rest, m)
	}
	out := make([]map[string]interface{}, 0, 1+len(rest))
	if len(systemParts) > 0 {
		out = append(out, map[string]interface{}{
			"role":    "system",
			"content": strings.Join(systemParts, "\n\n"),
		})
	}
	return append(out, rest...)
}

func rawMessageText(content interface{}) string {
	switch v := content.(type) {
	case string:
		return strings.TrimSpace(v)
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
