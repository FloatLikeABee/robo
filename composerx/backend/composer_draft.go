package main

import (
	"encoding/json"
	"html"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const maxComposerDraftHTMLBytes = 512 * 1024

var (
	reComposerStyleBlock = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reComposerScriptBlock = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reComposerBr          = regexp.MustCompile(`(?i)<br\s*/?>`)
	reComposerBlockClose  = regexp.MustCompile(`(?i)</(p|div|tr|h[1-6]|li|table|section|article|td)>`)
	reComposerTags        = regexp.MustCompile(`(?s)<[^>]+>`)
	reComposerHeading     = regexp.MustCompile(`(?is)<h([1-6])[^>]*>(.*?)</h[1-6]>`)
	reComposerLi          = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
)

func normalizeComposerDraft(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if len(s) > maxComposerDraftHTMLBytes {
		s = s[:maxComposerDraftHTMLBytes]
	}
	if !looksLikeHTML(s) {
		return strings.TrimSpace(s)
	}
	return htmlDraftToMarkdown(s)
}

var looksLikeHTMLPattern = regexp.MustCompile(`(?i)<[a-z][^>]*>`)

func looksLikeHTML(s string) bool {
	return strings.Contains(s, "<") && looksLikeHTMLPattern.MatchString(s)
}

func htmlDraftToMarkdown(s string) string {
	s = reComposerStyleBlock.ReplaceAllString(s, "")
	s = reComposerScriptBlock.ReplaceAllString(s, "")
	s = reComposerHeading.ReplaceAllStringFunc(s, func(m string) string {
		sub := reComposerHeading.FindStringSubmatch(m)
		if len(sub) < 3 {
			return "\n"
		}
		level, err := strconv.Atoi(sub[1])
		if err != nil || level < 1 {
			level = 1
		}
		if level > 6 {
			level = 6
		}
		text := strings.TrimSpace(stripInlineHTML(sub[2]))
		if text == "" {
			return "\n"
		}
		return "\n" + strings.Repeat("#", level) + " " + text + "\n"
	})
	s = reComposerLi.ReplaceAllStringFunc(s, func(m string) string {
		sub := reComposerLi.FindStringSubmatch(m)
		if len(sub) < 2 {
			return "\n"
		}
		text := strings.TrimSpace(stripInlineHTML(sub[1]))
		if text == "" {
			return "\n"
		}
		return "\n- " + text + "\n"
	})
	s = reComposerBr.ReplaceAllString(s, "\n")
	s = reComposerBlockClose.ReplaceAllString(s, "\n\n")
	s = reComposerTags.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return collapseBlankLines(s)
}

func stripInlineHTML(s string) string {
	s = reComposerTags.ReplaceAllString(s, "")
	return html.UnescapeString(strings.TrimSpace(s))
}

func collapseBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	prevBlank := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !prevBlank {
				out = append(out, "")
			}
			prevBlank = true
			continue
		}
		out = append(out, trimmed)
		prevBlank = false
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func safeParseComposerAIResponse(raw string) (msg string, prop *string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("composer-chat: draft parse panic: %v", r)
			msg = strings.TrimSpace(raw)
			prop = nil
		}
	}()
	return parseComposerAIResponse(raw)
}

func parseComposerAIResponse(raw string) (string, *string) {
	var msg string
	var prop *string

	var flex composerAIResponseFlexible
	if err := json.Unmarshal([]byte(raw), &flex); err == nil {
		msg, prop = mergeComposerAIFields(flex)
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err == nil {
		if prop == nil || strings.TrimSpace(derefString(prop)) == "" {
			for _, key := range []string{
				"proposed_markdown", "proposedMarkdown",
				"proposed_email_html", "proposedEmailHtml",
				"document", "markdown", "content", "draft", "body",
			} {
				if v, ok := m[key]; ok {
					if s, ok := v.(string); ok {
						if normalized := normalizeComposerDraft(s); normalized != "" {
							prop = &normalized
							break
						}
					}
				}
			}
		} else if normalized := normalizeComposerDraft(*prop); normalized != "" {
			prop = &normalized
		}
		if msg == "" {
			for _, key := range []string{"assistant_message", "assistantMessage"} {
				if v, ok := m[key]; ok {
					if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
						msg = strings.TrimSpace(s)
						break
					}
				}
			}
		}
	}

	if prop == nil {
		if extracted := extractMarkdownFence(msg); extracted != "" {
			prop = &extracted
		} else if extracted := extractMarkdownFence(raw); extracted != "" {
			prop = &extracted
		}
	}

	if msg == "" && prop == nil {
		return raw, nil
	}
	return msg, prop
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func extractMarkdownFence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	re := regexp.MustCompile("(?is)```(?:markdown|md)?\\s*([\\s\\S]*?)```")
	if m := re.FindStringSubmatch(text); len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}
