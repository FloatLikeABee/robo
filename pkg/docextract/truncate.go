package docextract

import "strings"

const (
	// MaxPerFileChars caps the text taken from a single file.
	MaxPerFileChars = 12000

	// MaxRequestChars caps the combined text across all files in one request.
	// Sized to leave room for the system prompt and instructions inside a
	// typical model context budget.
	MaxRequestChars = 24000

	// TruncationMarker is appended to any text that was cut, so both the model
	// and the user can tell content is missing.
	TruncationMarker = "\n\n…[truncated: content exceeded the size limit]"
)

// Truncate caps s at maxChars runes, appending TruncationMarker when it cuts.
// The second return value reports whether truncation occurred.
func Truncate(s string, maxChars int) (string, bool) {
	if maxChars <= 0 {
		return "", s != ""
	}
	r := []rune(s)
	if len(r) <= maxChars {
		return s, false
	}
	return strings.TrimSpace(string(r[:maxChars])) + TruncationMarker, true
}
