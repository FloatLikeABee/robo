package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
	"github.com/robo/webresearch"
)

type formTemplateAIRequest struct {
	Messages       []assistantMessage `json:"messages"`
	FormName       string             `json:"form_name"`
	FormDescription string            `json:"form_description"`
	CurrentHTML    string             `json:"current_html"`
	UseWebSearch   *bool              `json:"use_web_search"`
	WebSearchQuery string             `json:"web_search_query"`
}

type proposedFormQuestion struct {
	Title    string                 `json:"title"`
	Type     string                 `json:"type"`
	Required bool                   `json:"required"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type formTemplateAIResponse struct {
	AssistantMessage   string                 `json:"assistant_message"`
	ProposedFormHTML   *string                `json:"proposed_form_html,omitempty"`
	ProposedQuestions  []proposedFormQuestion `json:"proposed_questions,omitempty"`
	ResearchNotes      string                 `json:"research_notes,omitempty"`
	Sources            []webresearch.Source   `json:"sources,omitempty"`
}

type formTemplateAIResponseFlex struct {
	AssistantSnake  string                   `json:"assistant_message"`
	AssistantCamel  string                   `json:"assistantMessage"`
	ProposedSnake   *string                  `json:"proposed_form_html"`
	ProposedCamel   *string                  `json:"proposedFormHtml"`
	QuestionsSnake  []proposedFormQuestion   `json:"proposed_questions"`
	QuestionsCamel  []proposedFormQuestion   `json:"proposedQuestions"`
	ResearchNotes   string                   `json:"research_notes"`
	Sources         []webresearch.Source     `json:"sources"`
}

func mergeFormTemplateAIFields(f formTemplateAIResponseFlex) (msg string, html *string, questions []proposedFormQuestion) {
	msg = strings.TrimSpace(f.AssistantSnake)
	if msg == "" {
		msg = strings.TrimSpace(f.AssistantCamel)
	}
	if f.ProposedSnake != nil && strings.TrimSpace(*f.ProposedSnake) != "" {
		html = f.ProposedSnake
	} else if f.ProposedCamel != nil && strings.TrimSpace(*f.ProposedCamel) != "" {
		html = f.ProposedCamel
	}
	questions = f.QuestionsSnake
	if len(questions) == 0 {
		questions = f.QuestionsCamel
	}
	return msg, html, questions
}

func lastUserContent(messages []assistantMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return ""
}

func webSearchEnabled(useFlag *bool) bool {
	if useFlag == nil {
		return true
	}
	return *useFlag
}

func resolveWebSearchQuery(payload formTemplateAIRequest) string {
	if q := strings.TrimSpace(payload.WebSearchQuery); q != "" {
		return q
	}
	if q := lastUserContent(payload.Messages); q != "" {
		return q
	}
	parts := []string{strings.TrimSpace(payload.FormName), strings.TrimSpace(payload.FormDescription)}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// FormTemplateAIChat generates a web-grounded HTML form template and optional question suggestions.
func (h *Handler) FormTemplateAIChat(c *gin.Context) {
	if h.AI == nil || !h.AI.Configured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MorphAI is not configured (set MORPH_AI_API_KEY)"})
		return
	}

	var payload formTemplateAIRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if len(payload.Messages) == 0 && strings.TrimSpace(payload.FormName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages or form_name is required"})
		return
	}

	var researchNotes string
	var sources []webresearch.Source
	if webSearchEnabled(payload.UseWebSearch) {
		if q := resolveWebSearchQuery(payload); q != "" {
			researchNotes, sources = webresearch.Gather(q)
		}
	}

	system := `You are an expert form designer for SheetX.
CRITICAL: Output one JSON object only with exactly these keys:
- "assistant_message": concise plain-text guidance for the user.
- "proposed_form_html": complete HTML for a polished public form landing page (include embedded <style>; semantic form fields with name attributes matching field slugs), OR null if no HTML update is needed.
- "proposed_questions": array of suggested form fields OR empty array. Each item: {"title":"...","type":"text|select|multiselect|boolean|image|document","required":true/false,"config":{}}. For select/multiselect include config.options as [{"value":1,"label":"..."}].

Design goals:
- Modern, accessible, mobile-friendly layout.
- Use web research when provided; do not invent facts contradicting research.
- Map visible HTML inputs to proposed_questions titles/types.
- Avoid external scripts unless explicitly requested.`

	if name := strings.TrimSpace(payload.FormName); name != "" {
		system += "\n\nForm name: " + name
	}
	if desc := strings.TrimSpace(payload.FormDescription); desc != "" {
		system += "\n\nForm description: " + desc
	}
	if html := strings.TrimSpace(payload.CurrentHTML); html != "" {
		system += "\n\nCurrent saved HTML template (improve or replace as asked):\n" + html
	}
	if researchNotes != "" {
		system += "\n\nWeb research (ground truth for topic and common form fields):\n" + researchNotes
	}

	messages := []morphai.Message{{Role: "user", Content: system + "\n\nReply with JSON only."}}
	for _, m := range payload.Messages {
		role := strings.TrimSpace(m.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		messages = append(messages, morphai.Message{Role: role, Content: m.Content})
	}
	if len(payload.Messages) == 0 {
		messages = append(messages, morphai.Message{
			Role:    "user",
			Content: "Create a form HTML template and suggested questions based on the form name and description.",
		})
	}

	reply, err := h.AI.ChatCompletionLong(c.Request.Context(), messages)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("AI chat: %v", err)})
		return
	}
	reply = strings.TrimSpace(reply)
	if reply == "" {
		c.JSON(http.StatusBadGateway, gin.H{"error": "empty model response"})
		return
	}

	raw, ok := extractJSONObject(reply)
	if !ok {
		c.JSON(http.StatusOK, formTemplateAIResponse{AssistantMessage: reply, ResearchNotes: researchNotes, Sources: sources})
		return
	}

	var flex formTemplateAIResponseFlex
	if err := json.Unmarshal([]byte(raw), &flex); err != nil {
		c.JSON(http.StatusOK, formTemplateAIResponse{AssistantMessage: reply, ResearchNotes: researchNotes, Sources: sources})
		return
	}
	msg, html, questions := mergeFormTemplateAIFields(flex)
	if msg == "" && html != nil {
		msg = "Form template is ready. Apply the HTML and questions when you are happy."
	}
	if msg == "" {
		msg = reply
	}
	c.JSON(http.StatusOK, formTemplateAIResponse{
		AssistantMessage:  msg,
		ProposedFormHTML:  html,
		ProposedQuestions: questions,
		ResearchNotes:     researchNotes,
		Sources:           sources,
	})
}

// WebSearch returns lightweight public research notes for Morph AI tool loops.
func (h *Handler) WebSearch(c *gin.Context) {
	var body struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	query := strings.TrimSpace(body.Query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	notes, sources := webresearch.Gather(query)
	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"notes":   notes,
		"sources": sources,
	})
}

func execWebSearchTool(args map[string]interface{}) (string, error) {
	query := stringArg(args, "query")
	if query == "" {
		return "", fmt.Errorf("web_search requires query")
	}
	notes, sources := webresearch.Gather(query)
	payload := ginH{"query": query, "notes": notes, "sources": sources}
	return mustJSON(payload), nil
}
