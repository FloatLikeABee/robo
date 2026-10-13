package morphai

import (
	_ "embed"
	"strings"
	"unicode/utf8"
)

// HTMLSkillTokenBudget is the Ollama injection ceiling for the compact skill.
const HTMLSkillTokenBudget = 512

//go:embed htmlskill_condensed.md
var htmlSkillCondensed string

// EstimateTokens is Unicode runes / 4, rounding up.
func EstimateTokens(s string) int {
	return (utf8.RuneCountInString(s) + 3) / 4
}

// HTMLSkillInstructions is the compact html-anything skill plus Ge's overlay.
func HTMLSkillInstructions() string {
	return strings.TrimSpace(htmlSkillCondensed)
}

// WithHTMLSkill appends the compact skill after the caller's prompt so a title
// or seed stays first for small models.
func WithHTMLSkill(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	skill := HTMLSkillInstructions()
	if prompt == "" {
		return skill
	}
	if skill == "" {
		return prompt
	}
	return prompt + "\n\n" + skill
}
