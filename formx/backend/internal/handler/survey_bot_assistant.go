package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/robo/webresearch"
)

// SurveyBotRuntime is persisted in assistant conversation state.
type SurveyBotRuntime struct {
	TemplateID   string            `json:"template_id,omitempty"`
	TemplateSlug string            `json:"template_slug,omitempty"`
	Title        string            `json:"title,omitempty"`
	StepIndex    int               `json:"step_index"`
	Answers      map[string]string `json:"answers,omitempty"`
	Status       string            `json:"status,omitempty"` // idle|running|awaiting_ui|awaiting_create|creating_template|completed
	DraftMD      string            `json:"draft_md,omitempty"`
	CreateQuery  string            `json:"create_query,omitempty"`
}

type surveyBotTurn struct {
	Message  string
	State    assistantConversation
	UIBlocks []surveybot.UIBlock
	Record   any
	Done     bool
}

func mentionsSurveyBot(msg string) bool {
	return strings.Contains(strings.ToLower(msg), "survey bot")
}

func ensureSurveyBotState(st *assistantConversation) {
	if st.SurveyBot == nil {
		st.SurveyBot = &SurveyBotRuntime{
			Answers: map[string]string{},
			Status:  "idle",
		}
	}
	if st.SurveyBot.Answers == nil {
		st.SurveyBot.Answers = map[string]string{}
	}
	if st.Fields == nil {
		st.Fields = map[string]string{}
	}
}

func answersSummary(sb *SurveyBotRuntime) string {
	if sb == nil {
		return ""
	}
	return surveybot.FormatAnswersMarkdown(sb.Answers)
}

// runSurveyBotTurn handles one Survey Bot chat turn (deterministic engine + optional LLM tools later).
func (h *Handler) runSurveyBotTurn(ctx context.Context, req assistantChatRequest, lastUser string) surveyBotTurn {
	st := req.State
	ensureSurveyBotState(&st)
	st.Intent = "survey_bot"
	sb := st.SurveyBot
	h.ensureSurveyBotSeeds(ctx)

	// Structured UI widget answer
	if field, value, ok := surveybot.ParseSurveyAnswerMessage(lastUser); ok {
		return h.applySurveyAnswer(ctx, st, field, value)
	}

	low := strings.ToLower(strings.TrimSpace(lastUser))

	// Awaiting create confirmation
	if sb.Status == "awaiting_create" {
		if isYes(low) || strings.Contains(low, "create") {
			return h.beginTemplateCreation(ctx, st, sb.CreateQuery)
		}
		if isNo(low) {
			sb.Status = "idle"
			st.Intent = "general"
			st.SurveyBot = nil
			return surveyBotTurn{
				Message: "Okay — Survey Bot cancelled. Ask again with **survey bot** when you want to start.",
				State:   st,
				Done:    true,
			}
		}
		block := surveybot.ConfirmBlock("create_template", "Create a new Survey Bot template for this request?")
		return surveyBotTurn{
			Message:  "Should I create a new Survey Bot markdown template for your request?",
			State:    st,
			UIBlocks: []surveybot.UIBlock{block},
		}
	}

	// Awaiting draft approval
	if sb.Status == "creating_template" && sb.DraftMD != "" {
		if isYes(low) || strings.Contains(low, "save") || strings.Contains(low, "approve") {
			return h.saveDraftTemplateAndStart(ctx, st)
		}
		if isNo(low) || strings.Contains(low, "cancel") {
			sb.Status = "idle"
			sb.DraftMD = ""
			return surveyBotTurn{
				Message: "Draft discarded. Say **survey bot** plus what you need, or open the Survey Bot module to edit templates.",
				State:   st,
				Done:    true,
			}
		}
		block := surveybot.ConfirmBlock("approve_draft", "Save this draft template and start the survey?")
		return surveyBotTurn{
			Message:  "Review the draft below. Approve to save & start, or say **cancel**.\n\n```markdown\n" + truncateStr(sb.DraftMD, 3500) + "\n```",
			State:    st,
			UIBlocks: []surveybot.UIBlock{block},
		}
	}

	// Active survey: free-text answer for current text step
	if sb.Status == "running" || sb.Status == "awaiting_ui" {
		if sb.TemplateID != "" {
			return h.continueSurveyWithUserText(ctx, st, lastUser)
		}
	}

	// New / matching flow
	query := stripSurveyBotTrigger(lastUser)
	if query == "" {
		query = "survey"
	}
	matches, _ := h.SurveyBotTemplateRepo.SearchRanked(ctx, query, 5)
	if len(matches) == 0 {
		// also try list all and keyword
		all, _, _ := h.SurveyBotTemplateRepo.List(ctx, "", 1, 50)
		for _, t := range all {
			if surveybot.SearchScore(t.Title, t.Summary, t.Tags, t.Markdown, query) > 0 {
				matches = append(matches, t)
			}
		}
	}
	if len(matches) == 0 {
		sb.Status = "awaiting_create"
		sb.CreateQuery = query
		block := surveybot.ConfirmBlock("create_template", "No matching template. Create one with AI?")
		return surveyBotTurn{
			Message:  fmt.Sprintf("I couldn't find a Survey Bot template close to **%s**.\n\nCreate one from web research?", query),
			State:    st,
			UIBlocks: []surveybot.UIBlock{block},
		}
	}

	// Pick best match
	best := matches[0]
	return h.startSurveyFromTemplate(ctx, st, &best)
}

func (h *Handler) beginTemplateCreation(ctx context.Context, st assistantConversation, query string) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot
	notes, _ := webresearch.Gather(query + " survey questionnaire")
	md := draftSurveyMarkdown(query, query, "Survey for: "+query, notes)
	sb.Status = "creating_template"
	sb.DraftMD = md
	sb.CreateQuery = query
	block := surveybot.ConfirmBlock("approve_draft", "Save this draft template and start the survey?")
	return surveyBotTurn{
		Message: "Here's a draft Survey Bot template (web-grounded heuristics). Approve to save & start:\n\n```markdown\n" +
			truncateStr(md, 3500) + "\n```\n\n" + truncateStr(notes, 800),
		State:    st,
		UIBlocks: []surveybot.UIBlock{block},
	}
}

func (h *Handler) saveDraftTemplateAndStart(ctx context.Context, st assistantConversation) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot
	parsed, err := surveybot.ParseMarkdown(sb.DraftMD)
	if err != nil {
		return surveyBotTurn{Message: "Draft invalid: " + err.Error(), State: st}
	}
	slug := parsed.Slug
	if slug == "" {
		slug = normalizeSlug(parsed.Title)
	}
	t := &models.SurveyBotTemplate{
		Slug:     slug,
		Title:    parsed.Title,
		Tags:     parsed.Tags,
		Markdown: sb.DraftMD,
		Summary:  truncateStr(parsed.Instructions, 160),
	}
	if err := h.SurveyBotTemplateRepo.Insert(ctx, t); err != nil {
		return surveyBotTurn{Message: "Could not save template: " + err.Error(), State: st}
	}
	sb.DraftMD = ""
	return h.startSurveyFromTemplate(ctx, st, t)
}

func (h *Handler) startSurveyFromTemplate(ctx context.Context, st assistantConversation, t *models.SurveyBotTemplate) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot
	parsed, err := surveybot.ParseMarkdown(t.Markdown)
	if err != nil {
		return surveyBotTurn{Message: "Template parse error: " + err.Error(), State: st}
	}
	if err := surveybot.RequireSteps(parsed); err != nil {
		return surveyBotTurn{Message: err.Error() + " Open **AI Sheets** to compile questions from the description.", State: st}
	}
	sb.TemplateID = t.ID
	sb.TemplateSlug = t.Slug
	sb.Title = t.Title
	sb.StepIndex = 0
	sb.Answers = map[string]string{}
	sb.Status = "running"
	return h.promptCurrentStep(st, parsed)
}

func (h *Handler) continueSurveyWithUserText(ctx context.Context, st assistantConversation, lastUser string) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot
	t, err := h.SurveyBotTemplateRepo.GetByID(ctx, sb.TemplateID)
	if err != nil {
		return surveyBotTurn{Message: "Lost template — say **survey bot** to restart.", State: st, Done: true}
	}
	parsed, err := surveybot.ParseMarkdown(t.Markdown)
	if err != nil {
		return surveyBotTurn{Message: err.Error(), State: st}
	}
	if sb.StepIndex < 0 || sb.StepIndex >= len(parsed.Steps) {
		return h.finalizeSurvey(ctx, st, parsed)
	}
	step := parsed.Steps[sb.StepIndex]
	if step.Collect == "mcp_html" {
		// Re-prompt widget if they typed instead of clicking
		return h.promptCurrentStep(st, parsed)
	}
	val := strings.TrimSpace(lastUser)
	if val == "" && step.Required {
		return surveyBotTurn{Message: step.Prompt + "\n\n" + answersSummary(sb), State: st}
	}
	sb.Answers[step.Field] = val
	sb.StepIndex++
	if sb.StepIndex >= len(parsed.Steps) {
		return h.finalizeSurvey(ctx, st, parsed)
	}
	return h.promptCurrentStep(st, parsed)
}

func (h *Handler) applySurveyAnswer(ctx context.Context, st assistantConversation, field, value string) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot

	// create / approve confirms
	if field == "create_template" {
		if isYes(value) {
			return h.beginTemplateCreation(ctx, st, firstNonEmpty(sb.CreateQuery, "custom survey"))
		}
		sb.Status = "idle"
		st.SurveyBot = nil
		st.Intent = "general"
		return surveyBotTurn{Message: "Okay, not creating a template.", State: st, Done: true}
	}
	if field == "approve_draft" {
		if isYes(value) {
			return h.saveDraftTemplateAndStart(ctx, st)
		}
		sb.DraftMD = ""
		sb.Status = "idle"
		return surveyBotTurn{Message: "Draft discarded.", State: st, Done: true}
	}

	if sb.TemplateID == "" {
		return surveyBotTurn{Message: "No active survey. Say **survey bot** to start.", State: st}
	}
	t, err := h.SurveyBotTemplateRepo.GetByID(ctx, sb.TemplateID)
	if err != nil {
		return surveyBotTurn{Message: "Template missing.", State: st}
	}
	parsed, err := surveybot.ParseMarkdown(t.Markdown)
	if err != nil {
		return surveyBotTurn{Message: err.Error(), State: st}
	}
	sb.Answers[field] = value
	// advance to next unanswered / next index
	sb.StepIndex++
	// if field matched current step with different index, find step by field
	for i, s := range parsed.Steps {
		if s.Field == field {
			sb.StepIndex = i + 1
			break
		}
	}
	if sb.StepIndex >= len(parsed.Steps) {
		return h.finalizeSurvey(ctx, st, parsed)
	}
	return h.promptCurrentStep(st, parsed)
}

func (h *Handler) promptCurrentStep(st assistantConversation, parsed *surveybot.ParsedTemplate) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot
	if sb.StepIndex < 0 || sb.StepIndex >= len(parsed.Steps) {
		return surveyBotTurn{Message: "Survey complete.", State: st, Done: true}
	}
	step := parsed.Steps[sb.StepIndex]
	msg := fmt.Sprintf("**AI Sheet — %s** (step %d/%d)\n\n%s\n\n%s",
		sb.Title, sb.StepIndex+1, len(parsed.Steps), step.Prompt, answersSummary(sb))
	var blocks []surveybot.UIBlock
	if step.Collect == "mcp_html" {
		sb.Status = "awaiting_ui"
		if b := surveybot.BlockForStep(step); b != nil {
			blocks = append(blocks, *b)
		}
	} else {
		sb.Status = "running"
	}
	return surveyBotTurn{Message: msg, State: st, UIBlocks: blocks}
}

func (h *Handler) finalizeSurvey(ctx context.Context, st assistantConversation, parsed *surveybot.ParsedTemplate) surveyBotTurn {
	ensureSurveyBotState(&st)
	sb := st.SurveyBot
	html := surveybot.RenderThemedHTML(sb.Title, sb.Answers, parsed.Title)
	res := &models.SurveyBotResult{
		TemplateID:   sb.TemplateID,
		TemplateSlug: sb.TemplateSlug,
		Title:        sb.Title,
		Answers:      sb.Answers,
		HTML:         html,
	}
	if h.SurveyBotResultRepo != nil {
		_ = h.SurveyBotResultRepo.Insert(ctx, res)
	}
	if h.FormRepo != nil && res.ID != "" {
		h.FormRepo.EnqueueGraphSync(ctx, "formsx", "survey_result", res.ID, "upsert")
	}
	sb.Status = "completed"
	msg := fmt.Sprintf("**Survey complete — %s**\n\n%s\n\nSaved themed HTML result", sb.Title, answersSummary(sb))
	if res.ID != "" {
		msg += fmt.Sprintf(" `%s`. Open **AI Sheets** in SheetX to preview.", res.ID)
	} else {
		msg += "."
	}
	rec, _ := json.Marshal(models.SurveyBotResultToMap(res, false))
	var record any
	_ = json.Unmarshal(rec, &record)
	return surveyBotTurn{
		Message: msg,
		State:   st,
		Record:  record,
		Done:    true,
	}
}

func stripSurveyBotTrigger(msg string) string {
	low := strings.ToLower(msg)
	idx := strings.Index(low, "survey bot")
	if idx < 0 {
		return strings.TrimSpace(msg)
	}
	rest := strings.TrimSpace(msg[idx+len("survey bot"):])
	rest = strings.TrimLeft(rest, ":-—–, ")
	return strings.TrimSpace(rest)
}

func isYes(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "yes" || s == "y" || s == "ok" || s == "okay" || s == "sure" || s == "approve" || s == "save"
}

func isNo(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "no" || s == "n" || s == "cancel" || s == "nope"
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
