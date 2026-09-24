package secretbox

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// FromEnv builds a box from MORPH_SECRETS_KEY and optional MORPH_SECRETS_KEY_PREVIOUS.
// It reads the process environment only. Callers that load dotenv must do so first.
func FromEnv() (Box, error) {
	currentText := strings.TrimSpace(os.Getenv("MORPH_SECRETS_KEY"))
	if currentText == "" {
		return nil, ErrKeyMissing
	}
	current, err := decodeKey("MORPH_SECRETS_KEY", currentText)
	if err != nil {
		return nil, err
	}
	previousText := strings.TrimSpace(os.Getenv("MORPH_SECRETS_KEY_PREVIOUS"))
	var previous [][]byte
	if previousText != "" {
		for _, part := range strings.Split(previousText, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				return nil, fmt.Errorf("%w: MORPH_SECRETS_KEY_PREVIOUS contains an empty entry", ErrInvalidKey)
			}
			key, err := decodeKey("MORPH_SECRETS_KEY_PREVIOUS", part)
			if err != nil {
				return nil, err
			}
			previous = append(previous, key)
		}
	}
	return New(current, previous...)
}

func decodeKey(label, value string) ([]byte, error) {
	encoded := stripSpace(value)
	if encoded == "" || strings.ContainsAny(encoded, "-_") {
		return nil, fmt.Errorf("%w: %s must be 32 bytes of standard base64", ErrInvalidKey, label)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		raw, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil || len(raw) != 32 {
		return nil, fmt.Errorf("%w: %s must be 32 bytes of standard base64", ErrInvalidKey, label)
	}
	return raw, nil
}

func stripSpace(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s)
}
