package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
	"github.com/robo/webresearch"
)

type surveyBotCompileBody struct {
	Markdown    string `json:"markdown"`
	Description string `json:"description"`
	TitleHint   string `json:"title_hint"`
	UseWebSearch *bool `json:"use_web_search"`
}

// CompileSurveyBotTemplate POST /survey-bot/templates/compile
// Turns a description-only MD/TXT brief into a full question template (with optional web search).
func (h *Handler) CompileSurveyBotTemplate(c *gin.Context) {
	var body surveyBotCompileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	src := strings.TrimSpace(body.Markdown)
	if src == "" {
		src = strings.TrimSpace(body.Description)
	}
	if src == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "markdown or description is required"})
		return
	}

	parsed, err := surveybot.ParseMarkdown(src)
	if err != nil {
		// Plain text without frontmatter — wrap as description sheet.
		wrapped := fmt.Sprintf("---\ntitle: %s\n---\n\n%s\n",
			firstNonEmpty(strings.TrimSpace(body.TitleHint), "AI Sheet"), src)
		parsed, err = surveybot.ParseMarkdown(wrapped)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		src = wrapped
	}

	if !parsed.NeedsCompile && len(parsed.Steps) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"markdown":          surveybot.FormatMarkdown(parsed),
			"compiled":          false,
			"assistant_message": "Questions already present — saved as-is.",
			"needs_compile":     false,
		})
		return
	}

	useWeb := true
	if body.UseWebSearch != nil {
		useWeb = *body.UseWebSearch
	}
	var notes string
	var sources []webresearch.Source
	if useWeb {
		q := firstNonEmpty(parsed.Title, body.TitleHint, parsed.Instructions)
		notes, sources = webresearch.Gather(q + " weekend activities ideas survey questions")
	}

	md, compileErr := h.compileSurveyMarkdown(c.Request.Context(), parsed, notes)
	if compileErr != nil {
		// Heuristic fallback
		md = draftSurveyMarkdown(parsed.Title, parsed.Title, parsed.Instructions, notes)
	}
	if _, err := surveybot.ParseMarkdown(md); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "compiled template invalid: " + err.Error(), "markdown": md})
		return
	}

	srcOut := make([]gin.H, 0, len(sources))
	for _, s := range sources {
		srcOut = append(srcOut, gin.H{"title": s.Title, "type": s.Type, "url": s.URL})
	}
	c.JSON(http.StatusOK, gin.H{
		"markdown":          md,
		"compiled":          true,
		"research_notes":    notes,
		"sources":           srcOut,
		"assistant_message": "Designed survey questions from the sheet description. Review, edit, then publish.",
		"needs_compile":     false,
	})
}

func (h *Handler) compileSurveyMarkdown(ctx context.Context, parsed *surveybot.ParsedTemplate, researchNotes string) (string, error) {
	if h.AI == nil {
		return "", fmt.Errorf("AI client not configured")
	}
	system := `You design SheetX AI Sheet survey templates.
Output ONLY a complete markdown template with YAML frontmatter and ## Qn sections.
Allowed collect modes:
- text (free text in chat)
- mcp_html with widget: select | multiselect | confirm (boolean yes/no)
For select/multiselect always include options: [A, B, C]
Keep 4–8 questions. Prefer text for open answers; use mcp_html for choices and yes/no.
Do not invent facts that contradict research notes.`

	user := fmt.Sprintf("Title: %s\nSlug: %s\nAdmin description / instructions:\n%s\n",
		parsed.Title, parsed.Slug, parsed.Instructions)
	if researchNotes != "" {
		user += "\nWeb research notes:\n" + researchNotes
	}
	user += "\nExample format:\n" + draftSurveyMarkdown(parsed.Title, parsed.Title, parsed.Instructions, "")

	reply, err := h.AI.ChatCompletionLong(ctx, []morphai.Message{
		{Role: "user", Content: system + "\n\n" + user},
	})
	if err != nil {
		return "", err
	}
	md := extractMarkdownFence(reply)
	if md == "" {
		md = strings.TrimSpace(reply)
	}
	if _, err := surveybot.ParseMarkdown(md); err != nil {
		return "", err
	}
	return md, nil
}

func extractMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		rest := s[i+3:]
		rest = strings.TrimPrefix(rest, "markdown")
		rest = strings.TrimPrefix(rest, "md")
		rest = strings.TrimLeft(rest, "\n")
		if j := strings.Index(rest, "```"); j >= 0 {
			return strings.TrimSpace(rest[:j])
		}
	}
	return ""
}

// CompileAndSaveSurveyBotTemplate POST /survey-bot/templates/:id/compile — compile and persist.
func (h *Handler) CompileAndSaveSurveyBotTemplate(c *gin.Context) {
	if h.SurveyBotTemplateRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	t, err := h.SurveyBotTemplateRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	parsed, err := surveybot.ParseMarkdown(t.Markdown)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !parsed.NeedsCompile && len(parsed.Steps) > 0 {
		c.JSON(http.StatusOK, models.SurveyBotTemplateToMap(t))
		return
	}
	notes, _ := webresearch.Gather(parsed.Title + " " + parsed.Instructions + " survey questions")
	md, err := h.compileSurveyMarkdown(c.Request.Context(), parsed, notes)
	if err != nil {
		md = draftSurveyMarkdown(parsed.Title, parsed.Title, parsed.Instructions, notes)
	}
	if _, err := surveybot.ParseMarkdown(md); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	t.Markdown = md
	if err := h.SurveyBotTemplateRepo.Update(c.Request.Context(), t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := models.SurveyBotTemplateToMap(t)
	out["compiled"] = true
	c.JSON(http.StatusOK, out)
}
