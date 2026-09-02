package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/formsx/backend/internal/models"
	"github.com/formsx/backend/internal/mongo"
	"github.com/formsx/backend/internal/surveybot"
	"github.com/gin-gonic/gin"
	"github.com/robo/morphai"
)

const (
	infoSheetEventSource          = "info_sheet"
	infoSheetEventAITimeout       = 45 * time.Second
	infoSheetEventSummarizeSystem = `You summarize a completed Info Sheet (field collection) into one Events & Info record.

Return ONLY a JSON object (no markdown fences, not an array):
{
  "title": "short concrete title",
  "detail": "markdown body summarizing the collected answers",
  "reporter": "name or role if stated in the answers, otherwise empty string",
  "time": "RFC3339 timestamp if a time is implied, otherwise empty string"
}

Rules:
- Use only the sheet title and answers. Do not invent people, sites, or numbers.
- title is required and must be specific (not "Event" or "Note").
- detail may use Markdown. Prefer the source wording.
- Leave time empty when no clock time is given.`
)

// RecordSurveyBotResultAsEvent POST /survey-bot/results/:id/record-event
//
// Retries Events & Info recording for a completed Info Sheet result.
func (h *Handler) RecordSurveyBotResultAsEvent(c *gin.Context) {
	if h.SurveyBotResultRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "survey bot not configured"})
		return
	}
	res, err := h.SurveyBotResultRepo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "result not found"})
		return
	}
	h.recordInfoSheetResultAsEvent(c.Request.Context(), res)
	updated, err := h.SurveyBotResultRepo.GetByID(c.Request.Context(), res.ID)
	if err != nil {
		c.JSON(http.StatusOK, models.SurveyBotResultToMap(res, false))
		return
	}
	c.JSON(http.StatusOK, models.SurveyBotResultToMap(updated, false))
}

func (h *Handler) recordInfoSheetResultAsEvent(ctx context.Context, res *models.SurveyBotResult) {
	if res == nil || res.ID == "" || h.EventInfoRepo == nil || h.SurveyBotResultRepo == nil {
		return
	}
	if res.EventRecorded && res.EventID != "" {
		return
	}
	if existing, err := h.EventInfoRepo.FindBySourceResultID(ctx, res.ID); err == nil && existing != nil {
		res.EventRecorded = true
		res.EventID = existing.ID
		res.EventRecordError = ""
		_ = h.SurveyBotResultRepo.Update(ctx, res)
		return
	}

	draft := fallbackInfoSheetEventDraft(res)
	if h.AI != nil && h.AI.Configured() {
		aiCtx, cancel := context.WithTimeout(ctx, infoSheetEventAITimeout)
		defer cancel()
		if summarized, err := h.summarizeInfoSheetAsEvent(aiCtx, res); err == nil {
			draft = summarized
		}
	}
	ev, err := h.insertInfoSheetEvent(ctx, res, draft)
	if err != nil {
		res.EventRecorded = false
		res.EventRecordError = err.Error()
		_ = h.SurveyBotResultRepo.Update(ctx, res)
		return
	}
	res.EventRecorded = true
	res.EventID = ev.ID
	res.EventRecordError = ""
	_ = h.SurveyBotResultRepo.Update(ctx, res)
}

func (h *Handler) summarizeInfoSheetAsEvent(ctx context.Context, res *models.SurveyBotResult) (eventInfoIngestDraft, error) {
	userPrompt := fmt.Sprintf(
		"Summarize this completed Info Sheet into one Events & Info record.\nCurrent UTC time for reference: %s\n\nTitle: %s\nSlug: %s\n\nAnswers:\n%s",
		time.Now().UTC().Format(time.RFC3339),
		strings.TrimSpace(res.Title),
		strings.TrimSpace(res.TemplateSlug),
		surveybot.FormatAnswersMarkdown(res.Answers),
	)
	reply, err := h.AI.ChatCompletion(ctx, []morphai.Message{
		{Role: "system", Content: infoSheetEventSummarizeSystem},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		return eventInfoIngestDraft{}, fmt.Errorf("AI request failed: %w", err)
	}
	return firstInfoSheetEventDraft(reply)
}

func fallbackInfoSheetEventDraft(res *models.SurveyBotResult) eventInfoIngestDraft {
	if res == nil {
		return eventInfoIngestDraft{
			Title:  "Info Sheet response",
			Detail: "Completed Info Sheet with no answers recorded.",
		}
	}
	title := strings.TrimSpace(res.Title)
	if title == "" {
		title = "Info Sheet response"
	}
	detail := strings.TrimSpace(surveybot.FormatAnswersMarkdown(res.Answers))
	if detail == "" {
		detail = "Completed Info Sheet with no answers recorded."
	}
	return eventInfoIngestDraft{Title: title, Detail: detail}
}

func firstInfoSheetEventDraft(raw string) (eventInfoIngestDraft, error) {
	extracted, ok := extractJSONArrayOrObject(raw)
	if !ok {
		return eventInfoIngestDraft{}, errors.New("AI reply was not valid JSON")
	}
	drafts, err := parseEventIngestDrafts(extracted)
	if err != nil {
		return eventInfoIngestDraft{}, err
	}
	if len(drafts) == 0 {
		return eventInfoIngestDraft{}, errors.New("AI returned no usable event")
	}
	return drafts[0], nil
}

func (h *Handler) insertInfoSheetEvent(ctx context.Context, res *models.SurveyBotResult, draft eventInfoIngestDraft) (*models.EventInfo, error) {
	if h.EventInfoRepo == nil || res == nil || res.ID == "" {
		return nil, errors.New("event store is not configured")
	}
	if existing, err := h.EventInfoRepo.FindBySourceResultID(ctx, res.ID); err == nil && existing != nil {
		return existing, nil
	} else if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, err
	}
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		return nil, errors.New("event title is empty")
	}
	if len(title) > eventInfoMaxTitleLen {
		title = title[:eventInfoMaxTitleLen]
	}
	detail := strings.TrimSpace(draft.Detail)
	if len(detail) > eventInfoMaxDetailLen {
		detail = detail[:eventInfoMaxDetailLen]
	}
	reporter := strings.TrimSpace(draft.Reporter)
	if len(reporter) > eventInfoMaxReporterLen {
		reporter = reporter[:eventInfoMaxReporterLen]
	}
	ev := &models.EventInfo{
		Title:          title,
		Detail:         detail,
		Reporter:       reporter,
		Source:         infoSheetEventSource,
		SourceResultID: res.ID,
	}
	if draft.Time != "" {
		if t, err := time.Parse(time.RFC3339, draft.Time); err == nil {
			ev.EventTime = t.UTC()
		} else if t, err := time.Parse(time.RFC3339Nano, draft.Time); err == nil {
			ev.EventTime = t.UTC()
		}
	}
	if ev.EventTime.IsZero() {
		ev.EventTime = time.Now().UTC()
	}
	if err := h.EventInfoRepo.Insert(ctx, ev); err != nil {
		return nil, err
	}
	return ev, nil
}
