package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robo/assistmd"
)

type platformAssistantRequest struct {
	Messages []platformAssistantMessage `json:"messages"`
	State    platformAssistantState     `json:"state"`
}

type platformAssistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type platformAssistantState struct {
	Intent string            `json:"intent"`
	Fields map[string]string `json:"fields"`
}

type platformAssistantResponse struct {
	AssistantMessage string                 `json:"assistant_message"`
	Intent           string                 `json:"intent,omitempty"`
	MissingFields    []string               `json:"missing_fields,omitempty"`
	State            platformAssistantState `json:"state"`
	Completed        bool                   `json:"completed"`
	Record           any                    `json:"record,omitempty"`
}

var platformEmailPattern = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)

func (a *App) platformAssistantChat(c *gin.Context) {
	var req platformAssistantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	if req.State.Fields == nil {
		req.State.Fields = map[string]string{}
	}
	userMsg := latestPlatformUserMessage(req.Messages)
	if userMsg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages must include at least one user message"})
		return
	}
	reconcilePlatformIntent(&req.State, userMsg)
	if req.State.Intent == "" {
		req.State.Intent = detectPlatformIntent(userMsg)
	}
	updatePlatformFields(req.State.Fields, userMsg)

	switch req.State.Intent {
	case "create_template":
		a.platformAssistantCreateTemplate(c, req.State)
		return
	case "list_templates":
		items, total, err := a.templates.List(c.Request.Context(), 10, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list templates"})
			return
		}
		msg := assistmd.Empty("No templates yet.")
		if len(items) > 0 {
			rows := make([]string, 0, len(items))
			for _, t := range items {
				tag := strings.TrimSpace(t.Tag)
				line := assistmd.NamedID(t.Name, t.ID)
				if tag != "" {
					line += fmt.Sprintf(" · tag `%s`", tag)
				}
				rows = append(rows, line)
			}
			msg = assistmd.BulletList(fmt.Sprintf("**%d template(s)**", total), rows)
		}
		c.JSON(http.StatusOK, platformAssistantResponse{
			AssistantMessage: msg,
			Intent:           "list_templates",
			State:            platformAssistantState{Fields: map[string]string{}},
			Completed:        true,
			Record: gin.H{
				"items": items,
				"total": total,
			},
		})
		return
	default:
		a.respondPlatformGeneralAssistant(c, req, userMsg)
	}
}

func (a *App) platformAssistantCreateTemplate(c *gin.Context, st platformAssistantState) {
	required := []string{"name"}
	missing := missingPlatformFields(st.Fields, required)
	if len(missing) > 0 {
		c.JSON(http.StatusOK, platformAssistantResponse{
			AssistantMessage: "I can create the template. Please provide: " + strings.Join(missing, ", "),
			Intent:           st.Intent,
			MissingFields:    missing,
			State:            st,
		})
		return
	}
	htmlContent := strings.TrimSpace(st.Fields["html_content"])
	if htmlContent == "" {
		htmlContent = "<p>Hello {{first_name}},</p><p>This is your TranMail draft.</p>"
	}
	id, err := a.templates.Create(
		c.Request.Context(),
		strings.TrimSpace(st.Fields["name"]),
		strings.TrimSpace(st.Fields["tag"]),
		strings.TrimSpace(st.Fields["description"]),
		htmlContent,
		1,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create template"})
		return
	}
	c.JSON(http.StatusOK, platformAssistantResponse{
		AssistantMessage: assistmd.Success("Template created", strings.TrimSpace(st.Fields["name"])),
		Intent:           st.Intent,
		State:            st,
		Completed:        true,
		Record: gin.H{
			"id": id,
		},
	})
}

func latestPlatformUserMessage(messages []platformAssistantMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return ""
}

func detectPlatformIntent(message string) string {
	low := strings.ToLower(message)
	switch {
	case strings.Contains(low, "create template"), strings.Contains(low, "new template"):
		return "create_template"
	case strings.Contains(low, "list templates"), strings.Contains(low, "show templates"):
		return "list_templates"
	default:
		return "general"
	}
}

// reconcilePlatformIntent clears stale create flows when the user asks for a list or general help.
func reconcilePlatformIntent(st *platformAssistantState, message string) {
	low := strings.ToLower(strings.TrimSpace(message))
	if st.Fields == nil {
		st.Fields = map[string]string{}
	}
	switch {
	case strings.Contains(low, "list templates"), strings.Contains(low, "show templates"):
		st.Intent = "list_templates"
		st.Fields = map[string]string{}
	case strings.Contains(low, "create template"), strings.Contains(low, "new template"):
		st.Intent = "create_template"
	case st.Intent == "create_template":
		if strings.Contains(low, "list ") || strings.Contains(low, "show ") ||
			strings.Contains(low, "help me write") || strings.Contains(low, "follow-up") {
			st.Intent = "general"
			st.Fields = map[string]string{}
		}
	}
}

func updatePlatformFields(fields map[string]string, message string) {
	setPlatformField(fields, "name", message)
	setPlatformField(fields, "tag", message)
	setPlatformField(fields, "description", message)
	setPlatformField(fields, "html_content", message)
	setPlatformField(fields, "markdown_content", message)
}

func setPlatformField(fields map[string]string, token, message string) {
	if v := extractPlatformValue(message, token); v != "" {
		fields[token] = v
	}
}

func extractPlatformValue(message, token string) string {
	low := strings.ToLower(message)
	idx := strings.Index(low, token)
	if idx < 0 {
		return ""
	}
	sub := strings.TrimSpace(message[idx+len(token):])
	sub = strings.TrimSpace(strings.TrimLeft(sub, ":=- "))
	if sub == "" {
		return ""
	}
	if cut := strings.IndexAny(sub, ",;\n"); cut > 0 {
		sub = sub[:cut]
	}
	return strings.TrimSpace(sub)
}

func missingPlatformFields(fields map[string]string, required []string) []string {
	var missing []string
	for _, field := range required {
		if strings.TrimSpace(fields[field]) == "" {
			missing = append(missing, field)
		}
	}
	return missing
}
