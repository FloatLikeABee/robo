package ai

import (
	_ "embed"
	"regexp"
	"strings"
)

//go:embed prompts/learning_guide.md
var learningGuideMarkdown string

// LearningGuidePrompt returns the markdown instructions for learning-guide responses.
func LearningGuidePrompt() string {
	return strings.TrimSpace(learningGuideMarkdown)
}

var learningGuideIntentPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(how\s+(can|do|should|to)\s+i\s+)?learn\b`),
	regexp.MustCompile(`(?i)\blearning\s+(guide|path|roadmap|plan)\b`),
	regexp.MustCompile(`(?i)\bstudy\s+(guide|plan|path|roadmap)\b`),
	regexp.MustCompile(`(?i)\b(get\s+started|start\s+learning)\b`),
	regexp.MustCompile(`(?i)\bguide\s+me\b`),
	regexp.MustCompile(`(?i)\broadmap\s+(for|to)\b`),
	regexp.MustCompile(`(?i)\bhow\s+to\s+study\b`),
	regexp.MustCompile(`(?i)\bbest\s+way\s+to\s+learn\b`),
}

// WantsLearningGuide is true when the user is asking how to learn a topic (chat, no attached docs).
func WantsLearningGuide(req ChatRequest) bool {
	if req.DocumentMode {
		return false
	}
	if len(req.DocIDs) > 0 {
		return false
	}
	if req.HelpYouLearn {
		return true
	}
	return matchesLearningGuideIntent(lastUserSnippet(req))
}

func matchesLearningGuideIntent(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}
	for _, re := range learningGuideIntentPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}

// LearningGuideResearchQuery builds a web-search query for common learning approaches.
func LearningGuideResearchQuery(userMsg string) string {
	topic := strings.TrimSpace(userMsg)
	if topic == "" {
		topic = "academic study skills"
	}
	if len(topic) > 280 {
		topic = topic[:280]
	}
	return "how to learn " + topic + " study methods roadmap common approach best practices"
}
