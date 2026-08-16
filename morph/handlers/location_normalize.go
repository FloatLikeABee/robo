package handlers

import (
	"encoding/json"
	"strings"
)

// normalizeLocationValue coerces API input into a JSON string for MySQL JSON columns
// (plain text becomes {"location":"..."}); returns nil for empty input.
func normalizeLocationValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, nil
		}
		if json.Valid([]byte(s)) {
			return s, nil
		}
		b, err := json.Marshal(map[string]interface{}{"location": s})
		if err != nil {
			return nil, err
		}
		return string(b), nil
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	}
}
