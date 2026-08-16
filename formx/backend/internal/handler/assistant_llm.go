package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/robo/morphai"
)

const formsXAssistantInstructions = `You are SheetX AI for the SheetX form/sheet builder.

Help staff design forms, add questions, manage events, and explain features.

Fast source selection:
- Use the fastest grounded tool before reasoning from memory.
- For SheetX system knowledge or form structure questions, use search_system_documents/list_system_documents, then get_system_document for detail.
- For Events & Info operations, use list_events first, then get_event_context/get_event for detail.
- Use web_search only when public design context or external examples are needed.

Question types: text, select, multiselect, boolean, image, document.

SELECT / MULTISELECT (required for dropdowns):
- type must be "select" (single dropdown) or "multiselect" (checkboxes).
- ALWAYS pass config.options with at least 2 choices.
- Preferred shape: {"options":[{"value":1,"label":"Option A"},{"value":2,"label":"Option B"}]}
- Shorthand also works: {"options":["Option A","Option B"]} (values auto-numbered 1,2,3…).

When building a form: create_form → create_question per field (create_page for multi-page). Confirm with /f/{slug}.

For tools, reply with ONLY one JSON object: {"tool":"<name>","args":{...}}

Tools:
- list_forms — page, limit, search (optional)
- get_form — id OR slug (includes questions with options)
- create_form — name, slug (required), description (optional)
- list_pages / create_page — form_id (+ name, sort_order)
- list_questions / create_question / update_question — form_id, title, type; config with options for select/multiselect
- update_question — question_id (required), form_id (required); optional title, type, required, page_id, sort_order, config
- list_events / get_event / get_event_context / create_event
- sync_system_documents / list_system_documents / search_system_documents / get_system_document
- web_search — query (required): lightweight public web research for form design
- list_survey_results — page, limit, search (optional): saved Survey Bot HTML results
- get_survey_result — id (required): one Survey Bot result including answers
- list_survey_templates / get_survey_template — Survey Bot markdown templates
- Tip: when the user says "survey bot …", the dedicated Survey Bot mode runs automatically.

After TOOL_RESULT: summarize in markdown. Another tool → JSON only. No tool → markdown only.
Do not invent ids. List before detail lookups.`

const formsXToolMaxRounds = 8
const formsXMaxToolResultRunes = morphai.DefaultToolResultMaxRunes

type formsXToolCall struct {
	Tool string                 `json:"tool"`
	Args map[string]interface{} `json:"args"`
}

func (h *Handler) chatWithFormsXLLM(ctx context.Context, req assistantChatRequest, lastUser string) (string, any, error) {
	if h.AI == nil || !h.AI.Configured() {
		return "", nil, fmt.Errorf("MorphAI not configured")
	}

	first := formsXAssistantInstructions
	if hist := formatAssistantHistory(req.Messages); hist != "" {
		first += "\n\nRecent conversation:\n" + hist
	}
	if st := formatAssistantState(req.State); st != "" {
		first += "\n\nActive assistant state:\n" + st
	}
	first += "\n\nLatest user message:\n" + lastUser

	messages := []morphai.Message{{Role: "user", Content: first}}
	var lastRecord any

	for round := 0; round < formsXToolMaxRounds; round++ {
		reply, err := h.AI.ChatCompletion(ctx, messages)
		if err != nil {
			return "", nil, err
		}
		reply = strings.TrimSpace(reply)
		if reply == "" {
			return "", nil, fmt.Errorf("empty model response")
		}

		obj, ok := morphai.ExtractJSONObject(reply)
		if !ok {
			return reply, lastRecord, nil
		}

		call, err := parseFormsXToolCall(obj)
		if err != nil {
			return reply, lastRecord, nil
		}

		result, record, execErr := h.execFormsXTool(ctx, call)
		if record != nil {
			lastRecord = record
		}
		toolMsg := "TOOL_RESULT"
		if execErr != nil {
			toolMsg += " error=" + execErr.Error()
		} else {
			toolMsg += "\n" + morphai.TruncateRunes(result, formsXMaxToolResultRunes)
		}

		followUp := morphai.ToolFollowUpPrompt(toolMsg)
		messages = append(messages, morphai.Message{Role: "user", Content: followUp})
	}

	return "I hit the tool limit while gathering data. Please narrow your question or try again.", lastRecord, nil
}

func formatAssistantHistory(messages []assistantMessage) string {
	start := 0
	if len(messages) > morphai.DefaultHistoryMaxMessages {
		start = len(messages) - morphai.DefaultHistoryMaxMessages
	}
	var b strings.Builder
	for _, m := range messages[start:] {
		role := strings.TrimSpace(m.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := morphai.TruncateHistoryContent(m.Content, morphai.DefaultHistoryMaxRunes)
		if content == "" {
			continue
		}
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(content)
		b.WriteByte('\n')
	}
	return strings.TrimSpace(b.String())
}

func formatAssistantState(st assistantConversation) string {
	if st.Intent == "" && len(st.Fields) == 0 {
		return ""
	}
	raw, _ := json.Marshal(st)
	return string(raw)
}

func parseFormsXToolCall(raw string) (*formsXToolCall, error) {
	var call formsXToolCall
	if err := json.Unmarshal([]byte(raw), &call); err != nil {
		return nil, err
	}
	call.Tool = strings.TrimSpace(call.Tool)
	if call.Tool == "" {
		return nil, fmt.Errorf("missing tool")
	}
	if call.Args == nil {
		call.Args = map[string]interface{}{}
	}
	return &call, nil
}

func (h *Handler) execFormsXTool(ctx context.Context, call *formsXToolCall) (string, any, error) {
	switch call.Tool {
	case "list_forms":
		page := intArg(call.Args, "page", 1)
		limit := intArg(call.Args, "limit", 10)
		if limit > 25 {
			limit = 25
		}
		search := stringArg(call.Args, "search")
		list, total, err := h.FormRepo.List(page, limit, search)
		if err != nil {
			return "", nil, err
		}
		compact := make([]ginH, 0, len(list))
		for _, f := range list {
			compact = append(compact, ginH{"id": f.ID, "name": f.Name, "slug": f.Slug})
		}
		payload := ginH{"forms": compact, "total": total, "page": page, "limit": limit}
		return mustJSON(payload), payload, nil

	case "get_form":
		if slug := stringArg(call.Args, "slug"); slug != "" {
			form, err := h.FormRepo.GetBySlug(slug)
			if err != nil {
				return "", nil, err
			}
			questions, err := h.QuestionRepo.ListByFormID(form.ID)
			if err != nil {
				return "", nil, err
			}
			pages, _ := h.PageRepo.ListByFormID(form.ID)
			payload := ginH{"form": compactFormSummary(form), "questions": compactQuestionsSummary(questions), "pages": pages}
			return mustJSON(payload), payload, nil
		}
		id := int64Arg(call.Args, "id")
		if id <= 0 {
			return "", nil, fmt.Errorf("get_form requires id or slug")
		}
		form, err := h.FormRepo.GetByID(id)
		if err != nil {
			return "", nil, err
		}
		questions, err := h.QuestionRepo.ListByFormID(form.ID)
		if err != nil {
			return "", nil, err
		}
		pages, _ := h.PageRepo.ListByFormID(form.ID)
		payload := ginH{"form": compactFormSummary(form), "questions": compactQuestionsSummary(questions), "pages": pages}
		return mustJSON(payload), payload, nil

	case "create_form":
		name := stringArg(call.Args, "name")
		if name == "" {
			return "", nil, fmt.Errorf("create_form requires name")
		}
		slug := normalizeSlug(stringArg(call.Args, "slug"))
		if slug == "" {
			slug = normalizeSlug(name)
		}
		if slug == "" {
			return "", nil, fmt.Errorf("create_form requires slug")
		}
		exists, err := h.FormRepo.ExistsSlug(slug, 0)
		if err != nil {
			return "", nil, err
		}
		if exists {
			return "", nil, fmt.Errorf("slug %q already in use", slug)
		}
		form := &models.Form{
			Name:        name,
			Description: stringArg(call.Args, "description"),
			Slug:        slug,
		}
		if err := h.FormRepo.Create(form); err != nil {
			return "", nil, err
		}
		if _, err := h.PageRepo.EnsureDefaultPage(form.ID); err != nil {
			return "", nil, err
		}
		payload := ginH{"form": form, "public_url": "/f/" + form.Slug}
		return mustJSON(payload), payload, nil

	case "list_pages":
		formID := int64Arg(call.Args, "form_id")
		if formID <= 0 {
			return "", nil, fmt.Errorf("list_pages requires form_id")
		}
		if _, err := h.PageRepo.EnsureDefaultPage(formID); err != nil {
			return "", nil, err
		}
		pages, err := h.PageRepo.ListByFormID(formID)
		if err != nil {
			return "", nil, err
		}
		payload := ginH{"pages": pages, "form_id": formID}
		return mustJSON(payload), payload, nil

	case "create_page":
		formID := int64Arg(call.Args, "form_id")
		if formID <= 0 {
			return "", nil, fmt.Errorf("create_page requires form_id")
		}
		if _, err := h.FormRepo.GetByID(formID); err != nil {
			return "", nil, err
		}
		existing, _ := h.PageRepo.ListByFormID(formID)
		sortOrder := intArg(call.Args, "sort_order", len(existing))
		page := &models.FormPage{
			FormID:    formID,
			Name:      strings.TrimSpace(stringArg(call.Args, "name")),
			SortOrder: sortOrder,
		}
		if err := h.PageRepo.Create(page); err != nil {
			return "", nil, err
		}
		return mustJSON(page), page, nil

	case "list_questions":
		formID := int64Arg(call.Args, "form_id")
		if formID <= 0 {
			return "", nil, fmt.Errorf("list_questions requires form_id")
		}
		questions, err := h.QuestionRepo.ListByFormID(formID)
		if err != nil {
			return "", nil, err
		}
		payload := ginH{"questions": compactQuestionsSummary(questions), "form_id": formID}
		return mustJSON(payload), payload, nil

	case "create_question":
		formID := int64Arg(call.Args, "form_id")
		if formID <= 0 {
			return "", nil, fmt.Errorf("create_question requires form_id")
		}
		title := stringArg(call.Args, "title")
		qType := stringArg(call.Args, "type")
		if title == "" || qType == "" {
			return "", nil, fmt.Errorf("create_question requires title and type")
		}
		if !models.ValidQuestionTypes[qType] {
			return "", nil, fmt.Errorf("invalid question type %q", qType)
		}
		pageID := int64Arg(call.Args, "page_id")
		if pageID == 0 {
			if _, err := h.PageRepo.EnsureDefaultPage(formID); err != nil {
				return "", nil, err
			}
			pages, err := h.PageRepo.ListByFormID(formID)
			if err != nil || len(pages) == 0 {
				return "", nil, fmt.Errorf("no pages for form")
			}
			pageID = pages[0].ID
		}
		config, err := parseQuestionConfigFromAI(call.Args["config"], qType)
		if err != nil {
			return "", nil, err
		}
		existing, _ := h.QuestionRepo.ListByFormID(formID)
		sortOrder := intArg(call.Args, "sort_order", len(existing))
		q := &models.Question{
			FormID:    formID,
			PageID:    pageID,
			Title:     title,
			Type:      qType,
			Required:  boolArg(call.Args, "required"),
			SortOrder: sortOrder,
			Config:    config,
		}
		if err := h.QuestionRepo.Create(q); err != nil {
			return "", nil, err
		}
		return mustJSON(compactQuestionSummary(*q)), q, nil

	case "update_question":
		formID := int64Arg(call.Args, "form_id")
		qID := int64Arg(call.Args, "question_id")
		if formID <= 0 || qID <= 0 {
			return "", nil, fmt.Errorf("update_question requires form_id and question_id")
		}
		q, err := h.QuestionRepo.GetByFormIDAndID(formID, qID)
		if err != nil {
			return "", nil, err
		}
		if title := stringArg(call.Args, "title"); title != "" {
			q.Title = title
		}
		if t := stringArg(call.Args, "type"); t != "" {
			if !models.ValidQuestionTypes[t] {
				return "", nil, fmt.Errorf("invalid question type %q", t)
			}
			q.Type = t
		}
		if _, ok := call.Args["required"]; ok {
			q.Required = boolArg(call.Args, "required")
		}
		if pageID := int64Arg(call.Args, "page_id"); pageID > 0 {
			if _, err := h.PageRepo.GetByFormIDAndID(formID, pageID); err != nil {
				return "", nil, fmt.Errorf("invalid page_id")
			}
			q.PageID = pageID
		}
		if call.Args["sort_order"] != nil {
			q.SortOrder = intArg(call.Args, "sort_order", q.SortOrder)
		}
		if raw, ok := call.Args["config"]; ok && raw != nil {
			config, err := parseQuestionConfigFromAI(raw, q.Type)
			if err != nil {
				return "", nil, err
			}
			q.Config = config
		}
		if err := h.QuestionRepo.Update(q); err != nil {
			return "", nil, err
		}
		return mustJSON(compactQuestionSummary(*q)), q, nil

	case "get_event_context":
		idHex := stringArg(call.Args, "event_id")
		if idHex == "" {
			return "", nil, fmt.Errorf("get_event_context requires event_id")
		}
		ev, err := h.EventInfoRepo.GetByID(ctx, idHex)
		if err != nil {
			return "", nil, err
		}
		payload := models.EventInfoAIContextResponse{
			Event: models.EventInfoToResponse(ev),
		}
		return mustJSON(payload), payload, nil

	case "get_event":
		idHex := stringArg(call.Args, "event_id")
		if idHex == "" {
			return "", nil, fmt.Errorf("get_event requires event_id")
		}
		ev, err := h.EventInfoRepo.GetByID(ctx, idHex)
		if err != nil {
			return "", nil, err
		}
		payload := models.EventInfoToResponse(ev)
		return mustJSON(payload), payload, nil

	case "list_events":
		page := intArg(call.Args, "page", 1)
		limit := intArg(call.Args, "limit", 30)
		list, total, err := h.EventInfoRepo.List(ctx, page, limit)
		if err != nil {
			return "", nil, err
		}
		out := make([]*models.EventInfoResponse, 0, len(list))
		for i := range list {
			out = append(out, models.EventInfoToResponse(&list[i]))
		}
		payload := ginH{"events": out, "total": total, "page": page, "limit": limit}
		return mustJSON(payload), payload, nil

	case "create_event":
		title := stringArg(call.Args, "title")
		if title == "" {
			return "", nil, fmt.Errorf("create_event requires title")
		}
		timeStr := stringArg(call.Args, "time")
		if timeStr == "" {
			timeStr = time.Now().UTC().Format(time.RFC3339)
		}
		eventTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return "", nil, fmt.Errorf("time must be RFC3339")
		}
		ev := &models.EventInfo{
			Title:     title,
			Detail:    stringArg(call.Args, "detail"),
			Reporter:  stringArg(call.Args, "reporter"),
			EventTime: eventTime.UTC(),
		}
		if err := h.EventInfoRepo.Insert(ctx, ev); err != nil {
			return "", nil, err
		}
		payload := models.EventInfoToResponse(ev)
		return mustJSON(payload), payload, nil

	case "sync_system_documents":
		if h.AIDocRepo == nil {
			return "", nil, fmt.Errorf("ai document repo is not configured")
		}
		payload, err := h.syncSystemDocuments(ctx)
		if err != nil {
			return "", nil, err
		}
		return mustJSON(payload), payload, nil

	case "list_system_documents":
		if h.AIDocRepo == nil {
			return "", nil, fmt.Errorf("ai document repo is not configured")
		}
		docType := stringArg(call.Args, "doc_type")
		search := stringArg(call.Args, "search")
		page := intArg(call.Args, "page", 1)
		limit := intArg(call.Args, "limit", 30)
		list, total, err := h.AIDocRepo.List(ctx, docType, search, page, limit)
		if err != nil {
			return "", nil, err
		}
		out := make([]ginH, 0, len(list))
		for i := range list {
			doc := &list[i]
			out = append(out, ginH{
				"id":       doc.ID,
				"doc_type": doc.DocType,
				"title":    doc.Title,
				"summary":  doc.Summary,
			})
		}
		payload := ginH{"documents": out, "total": total, "page": page, "limit": limit}
		return mustJSON(payload), payload, nil

	case "search_system_documents":
		if h.AIDocRepo == nil {
			return "", nil, fmt.Errorf("ai document repo is not configured")
		}
		query := stringArg(call.Args, "query")
		if query == "" {
			query = stringArg(call.Args, "search")
		}
		if query == "" {
			return "", nil, fmt.Errorf("search_system_documents requires query")
		}
		docType := stringArg(call.Args, "doc_type")
		page := intArg(call.Args, "page", 1)
		limit := intArg(call.Args, "limit", 25)
		if limit > 50 {
			limit = 50
		}
		list, total, err := h.AIDocRepo.List(ctx, docType, query, page, limit)
		if err != nil {
			return "", nil, err
		}
		out := make([]ginH, 0, len(list))
		for i := range list {
			doc := &list[i]
			out = append(out, ginH{
				"id":       doc.ID,
				"doc_type": doc.DocType,
				"title":    doc.Title,
				"summary":  doc.Summary,
				"tags":     doc.Tags,
			})
		}
		payload := ginH{"documents": out, "total": total, "page": page, "limit": limit, "query": query}
		return mustJSON(payload), payload, nil

	case "get_system_document":
		if h.AIDocRepo == nil {
			return "", nil, fmt.Errorf("ai document repo is not configured")
		}
		idHex := stringArg(call.Args, "id")
		if idHex == "" {
			return "", nil, fmt.Errorf("get_system_document requires id")
		}
		doc, err := h.AIDocRepo.GetByID(ctx, idHex)
		if err != nil {
			return "", nil, err
		}
		payload := models.AIDocumentToMap(doc)
		return mustJSON(payload), payload, nil

	case "web_search":
		result, err := execWebSearchTool(call.Args)
		if err != nil {
			return "", nil, err
		}
		return result, nil, nil

	case "list_survey_results":
		if h.SurveyBotResultRepo == nil {
			return "", nil, fmt.Errorf("survey bot not configured")
		}
		page := intArg(call.Args, "page", 1)
		limit := intArg(call.Args, "limit", 25)
		search := stringArg(call.Args, "search")
		list, total, err := h.SurveyBotResultRepo.List(ctx, search, page, limit)
		if err != nil {
			return "", nil, err
		}
		out := make([]ginH, 0, len(list))
		for i := range list {
			out = append(out, models.SurveyBotResultToMap(&list[i], false))
		}
		payload := ginH{"results": out, "total": total, "page": page, "limit": limit}
		return mustJSON(payload), payload, nil

	case "get_survey_result":
		if h.SurveyBotResultRepo == nil {
			return "", nil, fmt.Errorf("survey bot not configured")
		}
		idHex := stringArg(call.Args, "id")
		if idHex == "" {
			return "", nil, fmt.Errorf("get_survey_result requires id")
		}
		res, err := h.SurveyBotResultRepo.GetByID(ctx, idHex)
		if err != nil {
			return "", nil, err
		}
		payload := models.SurveyBotResultToMap(res, true)
		return mustJSON(payload), payload, nil

	case "list_survey_templates":
		h.ensureSurveyBotSeeds(ctx)
		if h.SurveyBotTemplateRepo == nil {
			return "", nil, fmt.Errorf("survey bot not configured")
		}
		page := intArg(call.Args, "page", 1)
		limit := intArg(call.Args, "limit", 25)
		search := stringArg(call.Args, "search")
		list, total, err := h.SurveyBotTemplateRepo.List(ctx, search, page, limit)
		if err != nil {
			return "", nil, err
		}
		out := make([]ginH, 0, len(list))
		for i := range list {
			out = append(out, models.SurveyBotTemplateToMap(&list[i]))
		}
		payload := ginH{"templates": out, "total": total, "page": page, "limit": limit}
		return mustJSON(payload), payload, nil

	case "get_survey_template":
		if h.SurveyBotTemplateRepo == nil {
			return "", nil, fmt.Errorf("survey bot not configured")
		}
		idHex := stringArg(call.Args, "id")
		if idHex == "" {
			idHex = stringArg(call.Args, "slug")
		}
		if idHex == "" {
			return "", nil, fmt.Errorf("get_survey_template requires id or slug")
		}
		t, err := h.SurveyBotTemplateRepo.GetByID(ctx, idHex)
		if err != nil {
			t2, err2 := h.SurveyBotTemplateRepo.GetBySlug(ctx, idHex)
			if err2 != nil {
				return "", nil, err
			}
			t = t2
		}
		payload := models.SurveyBotTemplateToMap(t)
		return mustJSON(payload), payload, nil

	default:
		return "", nil, fmt.Errorf("unknown tool %q", call.Tool)
	}
}

type ginH map[string]interface{}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

func intArg(args map[string]interface{}, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(n))
		if err == nil {
			return i
		}
	}
	return def
}

func int64Arg(args map[string]interface{}, key string) int64 {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(n), 10, 64)
		if err == nil {
			return i
		}
	}
	return 0
}

func stringArg(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	default:
		return strings.TrimSpace(fmt.Sprint(s))
	}
}

func boolArg(args map[string]interface{}, key string) bool {
	v, ok := args[key]
	if !ok {
		return false
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return strings.EqualFold(strings.TrimSpace(b), "true") || b == "1"
	case float64:
		return b != 0
	case int:
		return b != 0
	default:
		return false
	}
}

func compactFormSummary(form *models.Form) ginH {
	if form == nil {
		return nil
	}
	return ginH{
		"id":          form.ID,
		"name":        form.Name,
		"slug":        form.Slug,
		"description": form.Description,
	}
}

func compactQuestionsSummary(questions []models.Question) []ginH {
	out := make([]ginH, 0, len(questions))
	for i := range questions {
		out = append(out, compactQuestionSummary(questions[i]))
	}
	return out
}

func extractJSONObject(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") {
		return "", false
	}
	depth := 0
	for i, ch := range s {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}
	return "", false
}
