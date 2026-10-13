package handlers

import (
	"encoding/json"
	"errors"
)

var errNoModelJSON = errors.New("AI did not return JSON")

// parseModelJSON unmarshals a model reply that should be one JSON object.
// It strips fences, ignores surrounding prose, and escapes raw newlines/tabs
// inside strings so small local models can still parse.
func parseModelJSON(raw string, dest any) error {
	jsonStr, ok := extractJSONObjectAware(raw)
	if !ok {
		jsonStr, ok = extractJSONObject(raw)
	}
	if !ok {
		return errNoModelJSON
	}
	repaired := repairJSONStringControls(jsonStr)
	if err := json.Unmarshal([]byte(repaired), dest); err != nil {
		return err
	}
	return nil
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
