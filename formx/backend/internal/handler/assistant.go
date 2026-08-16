package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/assistmd"
)

type assistantChatRequest struct {
	Messages []assistantMessage    `json:"messages"`
	State    assistantConversation `json:"state"`
}

type assistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type assistantConversation struct {
	Intent    string            `json:"intent"`
	Fields    map[string]string `json:"fields"`
	SurveyBot *SurveyBotRuntime `json:"survey_bot,omitempty"`
}

type assistantChatResponse struct {
	AssistantMessage string                `json:"assistant_message"`
	Intent           string                `json:"intent,omitempty"`
	MissingFields    []string              `json:"missing_fields,omitempty"`
	State            assistantConversation `json:"state"`
	Completed        bool                  `json:"completed"`
	Record           any                   `json:"record,omitempty"`
	UIBlocks         []surveybot.UIBlock   `json:"ui_blocks,omitempty"`
}

var spacePattern = regexp.MustCompile(`\s+`)
var objectIDHexPattern = regexp.MustCompile(`\b[0-9a-f]{24}\b|\b[0-9a-f]{32}\b`)

// AssistantChat provides a MorphAI-compatible starter assistant for FormsX.
// It can answer general questions and safely run create flows for forms
// by collecting required fields through a clarification loop, and read Events & Info.
func (h *Handler) AssistantChat(c *gin.Context) {
	var req assistantChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.State.Fields == nil {
		req.State.Fields = map[string]string{}
	}

	lastUser := latestUserMessage(req.Messages)
	if lastUser == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages must include at least one user message"})
		return
	}

	reconcileAssistantIntent(&req.State, lastUser)
	updateAssistantFields(req.State.Fields, lastUser)

	// Survey Bot takes priority when triggered or already active.
	if mentionsSurveyBot(lastUser) || req.State.Intent == "survey_bot" ||
		(req.State.SurveyBot != nil && req.State.SurveyBot.Status != "" && req.State.SurveyBot.Status != "idle" && req.State.SurveyBot.Status != "completed") ||
		strings.HasPrefix(strings.ToLower(strings.TrimSpace(lastUser)), "survey_bot_answer:") {
		turn := h.runSurveyBotTurn(c.Request.Context(), req, lastUser)
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: turn.Message,
			Intent:           turn.State.Intent,
			State:            turn.State,
			Completed:        turn.Done,
			Record:           turn.Record,
			UIBlocks:         turn.UIBlocks,
		})
		return
	}

	// When MorphAI is configured, use the LLM + tools for almost everything.
	if h.AI != nil && h.AI.Configured() {
		h.respondGeneralAssistant(c, req, lastUser)
		return
	}

	switch req.State.Intent {
	case "create_form":
		h.handleAssistantCreateForm(c, req.State)
		return
	case "create_event":
		h.handleAssistantCreateEvent(c, req.State)
		return
	case "list_forms":
		list, total, err := h.FormRepo.List(1, 10, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		msg := assistmd.Empty("No forms yet.")
		if len(list) > 0 {
			items := make([]string, 0, len(list))
			for _, f := range list {
				items = append(items, assistmd.NamedSlug(f.Name, f.Slug))
			}
			msg = assistmd.BulletList(
				fmt.Sprintf("**%d form(s)**", total),
				items,
			)
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: msg,
			Intent:           "list_forms",
			State:            assistantConversation{Intent: "general", Fields: map[string]string{}},
			Completed:        true,
			Record: gin.H{
				"forms": list,
				"total": total,
			},
		})
		return
	case "list_events":
		list, total, err := h.EventInfoRepo.List(context.Background(), 1, 30)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out := make([]*models.EventInfoResponse, 0, len(list))
		for i := range list {
			out = append(out, models.EventInfoToResponse(&list[i]))
		}
		eventMsg := assistmd.Empty("No Events & Info entries yet.")
		if len(out) > 0 {
			items := make([]string, 0, len(out))
			for _, ev := range out {
				title := ev.Title
				if title == "" {
					title = "Untitled"
				}
				items = append(items, fmt.Sprintf("**%s** — `%s`", title, ev.ID))
			}
			eventMsg = assistmd.BulletList(
				fmt.Sprintf("**%d event(s)** — ask for context with `event_id: <id>`", total),
				items,
			)
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: eventMsg,
			Intent:           "list_events",
			State:            assistantConversation{Intent: "general", Fields: map[string]string{}},
			Completed:        true,
			Record: gin.H{
				"events": out,
				"total":  total,
			},
		})
		return
	case "event_ai_context":
		h.handleAssistantEventAIContext(c, req.State, lastUser)
		return
	default:
		h.respondGeneralAssistant(c, req, lastUser)
	}
}

func (h *Handler) respondGeneralAssistant(c *gin.Context, req assistantChatRequest, lastUser string) {
	if h.AI != nil && h.AI.Configured() {
		reply, record, err := h.chatWithFormsXLLM(c.Request.Context(), req, lastUser)
		if err == nil {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: reply,
				Intent:           "general",
				State:            assistantConversation{Intent: "general", Fields: map[string]string{}},
				Completed:        true,
				Record:           record,
			})
			return
		}
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: fmt.Sprintf(
				"The AI assistant could not reach the model:\n\n**%s**\n\nCheck `MORPH_AI_BASE_URL`, `MORPH_AI_MODEL`, and your API key in `formx/backend/.env`, then restart the backend (`./start-all.sh restart formx-api`).",
				err.Error(),
			),
			Intent:    "general",
			State:     assistantConversation{Intent: "general", Fields: map[string]string{}},
			Completed: true,
		})
		return
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: "I can help with **FormsX**: forms, Events & Info, and create flows.\n\nTry:\n- `list forms`\n- `list events`\n- `create form name: … slug: …`\n- `create event title: … time: …`\n\nSet `MORPH_AI_API_KEY` to enable the full LLM assistant.",
		Intent:           "general",
		State:            assistantConversation{Intent: "general", Fields: map[string]string{}},
		Completed:        true,
	})
}

func (h *Handler) handleAssistantEventAIContext(c *gin.Context, st assistantConversation, lastUser string) {
	idHex := strings.TrimSpace(st.Fields["event_id"])
	if idHex == "" {
		if x := extractObjectIDHex(lastUser); x != "" {
			idHex = x
			st.Fields["event_id"] = idHex
		}
	}
	if idHex == "" {
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: "Paste the event id or say event_id: <id> so I can load the event detail.",
			Intent:           st.Intent,
			MissingFields:    []string{"event_id"},
			State:            st,
		})
		return
	}
	ev, err := h.EventInfoRepo.GetByID(context.Background(), idHex)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusOK, assistantChatResponse{
				AssistantMessage: "No event found with that id.",
				Intent:           st.Intent,
				State:            st,
				Completed:        true,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	payload := models.EventInfoAIContextResponse{
		Event: models.EventInfoToResponse(ev),
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: "Loaded event detail.",
		Intent:           st.Intent,
		State:            st,
		Completed:        true,
		Record:           payload,
	})
}

func extractObjectIDHex(message string) string {
	m := objectIDHexPattern.FindString(strings.ToLower(message))
	return m
}

func (h *Handler) handleAssistantCreateForm(c *gin.Context, st assistantConversation) {
	required := []string{"name", "slug"}
	missing := missingAssistantFields(st.Fields, required)
	if len(missing) > 0 {
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: "I can create the form. Please provide: " + strings.Join(missing, ", "),
			Intent:           st.Intent,
			MissingFields:    missing,
			State:            st,
		})
		return
	}

	slug := normalizeSlug(st.Fields["slug"])
	if slug == "" {
		slug = normalizeSlug(st.Fields["name"])
	}
	exists, err := h.FormRepo.ExistsSlug(slug, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: "That slug is already in use. Please provide another slug.",
			Intent:           st.Intent,
			MissingFields:    []string{"slug"},
			State:            st,
		})
		return
	}

	form := &models.Form{
		Name:        st.Fields["name"],
		Description: st.Fields["description"],
		Slug:        slug,
	}
	if err := h.FormRepo.Create(form); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: assistmd.Success("Form created", fmt.Sprintf("`/%s`", form.Slug)),
		Intent:           st.Intent,
		State:            st,
		Completed:        true,
		Record:           form,
	})
}

func (h *Handler) handleAssistantCreateEvent(c *gin.Context, st assistantConversation) {
	required := []string{"title"}
	missing := missingAssistantFields(st.Fields, required)
	if len(missing) > 0 {
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: "I can log an Events & Info entry. Please provide: " + strings.Join(missing, ", ") + " (optional: detail, reporter, time as RFC3339 — defaults to now).",
			Intent:           st.Intent,
			MissingFields:    missing,
			State:            st,
		})
		return
	}

	timeStr := strings.TrimSpace(st.Fields["time"])
	if timeStr == "" {
		timeStr = time.Now().UTC().Format(time.RFC3339)
	}
	eventTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		c.JSON(http.StatusOK, assistantChatResponse{
			AssistantMessage: "The time must be RFC3339 / ISO-8601 (e.g. 2026-05-21T15:04:05Z). Please provide time: … again.",
			Intent:           st.Intent,
			MissingFields:    []string{"time"},
			State:            st,
		})
		return
	}

	ev := &models.EventInfo{
		Title:     strings.TrimSpace(st.Fields["title"]),
		Detail:    strings.TrimSpace(st.Fields["detail"]),
		Reporter:  strings.TrimSpace(st.Fields["reporter"]),
		EventTime: eventTime.UTC(),
	}
	if err := h.EventInfoRepo.Insert(context.Background(), ev); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := models.EventInfoToResponse(ev)
	c.JSON(http.StatusOK, assistantChatResponse{
		AssistantMessage: assistmd.Success("Event logged", fmt.Sprintf("**%s** — `%s`", resp.Title, resp.ID)),
		Intent:           st.Intent,
		State:            st,
		Completed:        true,
		Record:           resp,
	})
}

func latestUserMessage(messages []assistantMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return ""
}

// reconcileAssistantIntent applies the latest user message over stale client state.
// Without this, a prior create flow (e.g. from a suggestion chip) can stick and
// hijack unrelated requests like "List forms".
func reconcileAssistantIntent(st *assistantConversation, lastUser string) {
	detected := detectAssistantIntent(lastUser)
	switch detected {
	case "list_forms", "list_events", "event_ai_context":
		st.Intent = detected
		st.Fields = map[string]string{}
		st.SurveyBot = nil
	case "survey_bot":
		st.Intent = "survey_bot"
	case "create_form", "create_event":
		if st.Intent != detected {
			st.Fields = map[string]string{}
		}
		st.Intent = detected
		st.SurveyBot = nil
	case "general":
		switch st.Intent {
		case "create_form", "create_event":
			if !looksLikeCreateFieldReply(lastUser) {
				st.Intent = "general"
				st.Fields = map[string]string{}
			}
		case "survey_bot":
			// Keep active survey sessions across free-text answers.
			if st.SurveyBot != nil {
				switch st.SurveyBot.Status {
				case "running", "awaiting_ui", "awaiting_create", "creating_template":
					st.Intent = "survey_bot"
					return
				}
			}
			st.Intent = "general"
		default:
			st.Intent = "general"
			st.Fields = map[string]string{}
		}
	default:
		st.Intent = detected
		if st.Fields == nil {
			st.Fields = map[string]string{}
		}
	}
}

func looksLikeCreateFieldReply(message string) bool {
	low := strings.ToLower(message)
	tokens := []string{"name:", "slug:", "title:", "time:", "detail:", "reporter:", "description:"}
	for _, t := range tokens {
		if strings.Contains(low, t) {
			return true
		}
	}
	return extractObjectIDHex(message) != ""
}

func detectAssistantIntent(message string) string {
	low := strings.ToLower(message)
	switch {
	case strings.Contains(low, "survey bot") || strings.HasPrefix(strings.TrimSpace(low), "survey_bot_answer:"):
		return "survey_bot"
	case strings.Contains(low, "create form"), strings.Contains(low, "new form"):
		return "create_form"
	case strings.Contains(low, "create event"), strings.Contains(low, "new event"),
		strings.Contains(low, "log event"), strings.Contains(low, "add event"):
		return "create_event"
	case strings.Contains(low, "list forms"), strings.Contains(low, "show forms"):
		return "list_forms"
	case strings.Contains(low, "list events"), strings.Contains(low, "show events"), strings.Contains(low, "events & info"):
		return "list_events"
	case strings.Contains(low, "event context"), strings.Contains(low, "event detail"), strings.Contains(low, "events context"):
		return "event_ai_context"
	default:
		if extractObjectIDHex(message) != "" && (strings.Contains(low, "context") || strings.Contains(low, "detail") || strings.Contains(low, "who")) {
			return "event_ai_context"
		}
		return "general"
	}
}

func updateAssistantFields(fields map[string]string, message string) {
	setFieldIfPresent(fields, "name", message, "name")
	setFieldIfPresent(fields, "slug", message, "slug")
	setFieldIfPresent(fields, "description", message, "description")
	setFieldIfPresent(fields, "title", message, "title")
	setFieldIfPresent(fields, "detail", message, "detail")
	setFieldIfPresent(fields, "reporter", message, "reporter")
	setFieldIfPresent(fields, "time", message, "time")
	setFieldIfPresent(fields, "event_id", message, "event_id")

	if x := extractObjectIDHex(message); x != "" {
		fields["event_id"] = x
	}
}

func setFieldIfPresent(fields map[string]string, key, message, token string) {
	if v := extractTokenValue(message, token); v != "" {
		fields[key] = v
	}
}

func extractTokenValue(message, token string) string {
	low := strings.ToLower(message)
	idx := strings.Index(low, token)
	if idx < 0 {
		return ""
	}
	sub := message[idx+len(token):]
	sub = strings.TrimSpace(strings.TrimLeft(sub, ":=- "))
	if sub == "" {
		return ""
	}
	if cut := strings.IndexAny(sub, ",;\n"); cut > 0 {
		sub = sub[:cut]
	}
	return strings.TrimSpace(sub)
}

func missingAssistantFields(fields map[string]string, required []string) []string {
	out := make([]string, 0)
	for _, field := range required {
		if strings.TrimSpace(fields[field]) == "" {
			out = append(out, field)
		}
	}
	return out
}

func normalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "_", "-")
	s = spacePattern.ReplaceAllString(s, "-")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	out := strings.Trim(b.String(), "-")
	out = strings.ReplaceAll(out, "--", "-")
	return out
}
