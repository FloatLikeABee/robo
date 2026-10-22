package handlers

import "strings"

func replyLanguageLine(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "zh":
		return "Reply in Simplified Chinese unless the user writes in another language."
	case "en":
		return "Reply in English unless the user writes in another language."
	default:
		return ""
	}
}
