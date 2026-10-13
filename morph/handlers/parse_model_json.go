package handlers

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode"
)

var errNoModelJSON = errors.New("AI did not return JSON")

// parseModelJSON unmarshals a model reply that should be one JSON object.
// It strips fences, tries each balanced {...} block, escapes raw newlines/tabs
// inside strings, and inserts missing commas between properties.
func parseModelJSON(raw string, dest any) error {
	var last error
	for _, jsonStr := range extractJSONObjects(raw) {
		repaired := repairJSONPunctuation(repairJSONStringControls(jsonStr))
		if err := json.Unmarshal([]byte(repaired), dest); err != nil {
			last = err
			continue
		}
		return nil
	}
	if last != nil {
		return last
	}
	return errNoModelJSON
}

func stripModelFences(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func extractJSONObjects(s string) []string {
	s = stripModelFences(s)
	var out []string
	for i := 0; i < len(s); i++ {
		if s[i] != '{' {
			continue
		}
		obj, ok := balancedObjectAt(s, i)
		if !ok {
			continue
		}
		out = append(out, obj)
		i += len(obj) - 1
	}
	return out
}

func balancedObjectAt(s string, start int) (string, bool) {
	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

func repairJSONPunctuation(s string) string {
	var b []byte
	inString := false
	escaped := false
	expectValue := false
	valueDone := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inString {
			b = append(b, ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
				if expectValue {
					valueDone = true
					expectValue = false
				}
			}
			continue
		}
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
			b = append(b, ch)
			continue
		}
		if valueDone && (ch == '"' || ch == '{' || ch == '[') {
			b = append(b, ',')
			valueDone = false
		}
		switch ch {
		case '"':
			inString = true
			b = append(b, ch)
		case ':':
			expectValue = true
			valueDone = false
			b = append(b, ch)
		case ',':
			expectValue = false
			valueDone = false
			b = append(b, ch)
		case '{', '[':
			expectValue = false
			valueDone = false
			b = append(b, ch)
		case '}', ']':
			for len(b) > 0 && unicode.IsSpace(rune(b[len(b)-1])) {
				b = b[:len(b)-1]
			}
			if len(b) > 0 && b[len(b)-1] == ',' {
				b = b[:len(b)-1]
			}
			valueDone = true
			expectValue = false
			b = append(b, ch)
		default:
			j := i
			for j < len(s) {
				c := s[j]
				if c == ',' || c == '}' || c == ']' || c == ' ' || c == '\n' || c == '\r' || c == '\t' {
					break
				}
				j++
			}
			b = append(b, s[i:j]...)
			i = j - 1
			if expectValue {
				valueDone = true
				expectValue = false
			}
		}
	}
	return string(b)
}

func repairJSONStringControls(s string) string {
	var b []byte
	inString := false
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				b = append(b, ch)
				escaped = false
				continue
			}
			if ch == '\\' {
				b = append(b, ch)
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
				b = append(b, ch)
				continue
			}
			switch ch {
			case '\n':
				b = append(b, '\\', 'n')
			case '\r':
				b = append(b, '\\', 'r')
			case '\t':
				b = append(b, '\\', 't')
			default:
				b = append(b, ch)
			}
			continue
		}
		if ch == '"' {
			inString = true
		}
		b = append(b, ch)
	}
	return string(b)
}
