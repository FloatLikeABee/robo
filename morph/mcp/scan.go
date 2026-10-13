package mcp

import (
	"regexp"
	"strings"
)

// Finding is one secret or PII hit in a share snapshot.
type Finding struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

var (
	emailRe   = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)
	phoneRe   = regexp.MustCompile(`(?i)\b(?:\+?1[-.\s]?)?(?:\(?\d{3}\)?[-.\s]?)\d{3}[-.\s]?\d{4}\b`)
	jwtRe     = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)
	skRe      = regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9]{16,}\b`)
	awsRe     = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	keyEqRe   = regexp.MustCompile(`(?i)\b(?:api[_-]?key|secret|password|token|jwt_secret)\s*[:=]\s*\S+`)
	pemRe     = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)
	ipRe      = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	patRe     = regexp.MustCompile(`\bmorph_pat_[0-9a-f]+_[0-9a-f]+\b`)
)

// ScanShareText returns secret and PII findings. Empty means the snapshot is clean.
func ScanShareText(s string) []Finding {
	var out []Finding
	add := func(kind, label string, re *regexp.Regexp) {
		if re.FindString(s) != "" {
			out = append(out, Finding{Kind: kind, Label: label})
		}
	}
	add("secret", "jwt", jwtRe)
	add("secret", "api_key", skRe)
	add("secret", "aws_key", awsRe)
	add("secret", "env_secret", keyEqRe)
	add("secret", "private_key", pemRe)
	add("secret", "pat", patRe)
	add("pii", "email", emailRe)
	add("pii", "phone", phoneRe)
	add("pii", "ip", ipRe)
	return out
}

func shareBlocked(findings []Finding) bool {
	return len(findings) > 0
}

func stripProvenance(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSpace(strings.TrimPrefix(s, "[morph-mcp]"))
	s = strings.TrimSpace(strings.TrimPrefix(s, "source: morph-mcp"))
	return strings.TrimSpace(s)
}
