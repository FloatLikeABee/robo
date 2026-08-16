package handler

import (
	"encoding/csv"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/formsx/backend/internal/mail"
	"github.com/formsx/backend/internal/models"
	"github.com/gin-gonic/gin"
)

// ListResponses returns paginated responses for a form.
// @Summary List form responses
// @Tags responses
// @Produce json
// @Param id path int true "Form ID"
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Param since query int false "Since (unix ms)"
// @Param until query int false "Until (unix ms)"
// @Success 200 {object} models.ResponseListResult
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/responses [get]
func (h *Handler) ListResponses(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	var since, until *int64
	if s := c.Query("since"); s != "" {
		if v, e := strconv.ParseInt(s, 10, 64); e == nil {
			since = &v
		}
	}
	if u := c.Query("until"); u != "" {
		if v, e := strconv.ParseInt(u, 10, 64); e == nil {
			until = &v
		}
	}
	list, total, err := h.ResponseRepo.List(c.Request.Context(), formID, page, limit, since, until)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ResponseListResult{Responses: list, Total: total, Page: page, Limit: limit})
}

// ExportResponses exports responses as CSV.
// @Summary Export responses as CSV
// @Tags responses
// @Produce text/csv
// @Param id path int true "Form ID"
// @Success 200 file csv
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/responses/export [get]
func (h *Handler) ExportResponses(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	form, err := h.FormRepo.GetByID(formID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	questions, err := h.QuestionRepo.ListByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	list, _, err := h.ResponseRepo.List(c.Request.Context(), formID, 1, 10000, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Build CSV: header = Submitted At, then question titles
	titleByID := make(map[int64]string)
	for _, q := range questions {
		titleByID[q.ID] = q.Title
	}
	headers := []string{"Submitted At", "Exam duration (sec)"}
	for _, q := range questions {
		headers = append(headers, q.Title)
	}
	c.Header("Content-Disposition", "attachment; filename="+form.Slug+"-responses.csv")
	c.Header("Content-Type", "text/csv")
	w := csv.NewWriter(c.Writer)
	_ = w.Write(headers)
	for _, r := range list {
		examSec := ""
		if r.ExamDurationMs > 0 {
			examSec = fmt.Sprintf("%.3f", float64(r.ExamDurationMs)/1000.0)
		}
		row := []string{r.SubmittedAt.Format("2006-01-02 15:04:05"), examSec}
		answerByQID := make(map[int64]interface{})
		for _, a := range r.Answers {
			answerByQID[a.QuestionID] = a.Value
		}
		for _, q := range questions {
			v := answerByQID[q.ID]
			row = append(row, valueToString(v))
		}
		_ = w.Write(row)
	}
	w.Flush()
}

// DeleteResponse deletes an individual response for a form.
// @Summary Delete individual response
// @Tags responses
// @Param id path int true "Form ID"
// @Param responseId path string true "Response ID"
// @Success 204
// @Failure 400,404 {object} map[string]string
// @Router /api/v1/forms/{formId}/responses/{responseId} [delete]
func (h *Handler) DeleteResponse(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	if _, err := h.FormRepo.GetByID(formID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	responseID := c.Param("responseId")
	if responseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid responseId"})
		return
	}
	if err := h.ResponseRepo.DeleteByID(c.Request.Context(), formID, responseID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "response not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

type EmailResponseRequest struct {
	To      []string `json:"to" binding:"required"`
	Subject string   `json:"subject"`
}

// EmailResponse emails an individual response as an HTML sheet.
// @Summary Email individual response
// @Tags responses
// @Accept json
// @Produce json
// @Param id path int true "Form ID"
// @Param responseId path string true "Response ID"
// @Router /api/v1/forms/{formId}/responses/{responseId}/email [post]
func (h *Handler) EmailResponse(c *gin.Context) {
	formID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid formId"})
		return
	}
	form, err := h.FormRepo.GetByID(formID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "form not found"})
		return
	}
	responseID := c.Param("responseId")
	if responseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid responseId"})
		return
	}
	var req EmailResponseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.ResponseRepo.GetByID(c.Request.Context(), formID, responseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "response not found"})
		return
	}
	qs, err := h.QuestionRepo.ListByFormID(formID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	subject := strings.TrimSpace(req.Subject)
	if subject == "" {
		subject = fmt.Sprintf("FormsX response: %s", form.Name)
	}

	answerByQID := make(map[int64]interface{})
	for _, a := range resp.Answers {
		answerByQID[a.QuestionID] = a.Value
	}

	var b strings.Builder
	b.WriteString("<div style=\"font-family:system-ui,-apple-system,Segoe UI,Roboto,Arial,sans-serif; color:#0f172a;\">")
	b.WriteString("<h2 style=\"margin:0 0 4px 0; font-size:18px;\">")
	b.WriteString(html.EscapeString(form.Name))
	b.WriteString("</h2>")
	if strings.TrimSpace(form.Description) != "" {
		b.WriteString("<div style=\"color:#475569; font-size:12px; margin-bottom:10px;\">")
		b.WriteString(html.EscapeString(form.Description))
		b.WriteString("</div>")
	}
	b.WriteString("<div style=\"color:#64748b; font-size:12px; margin-bottom:14px;\">Submitting time: ")
	b.WriteString(html.EscapeString(resp.SubmittedAt.Format("2006-01-02 15:04:05")))
	b.WriteString("</div>")
	if resp.ExamDurationMs > 0 {
		sec := fmt.Sprintf("%.2f", float64(resp.ExamDurationMs)/1000.0)
		b.WriteString("<div style=\"color:#64748b; font-size:12px; margin-bottom:14px;\">Exam time recorded: ")
		b.WriteString(html.EscapeString(sec))
		b.WriteString(" seconds</div>")
	}
	b.WriteString("<ol style=\"padding-left:18px; margin:0;\">")
	for _, q := range qs {
		b.WriteString("<li style=\"margin:0 0 10px 0;\">")
		b.WriteString("<div style=\"font-weight:600; font-size:13px; margin-bottom:2px;\">")
		b.WriteString(html.EscapeString(q.Title))
		b.WriteString("</div>")
		b.WriteString("<div style=\"font-size:13px; color:#0f172a;\">")
		b.WriteString(html.EscapeString(valueToString(answerByQID[q.ID])))
		b.WriteString("</div>")
		b.WriteString("</li>")
	}
	b.WriteString("</ol></div>")

	if err := mail.SendHTML(h.Cfg, req.To, subject, b.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func valueToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case []interface{}:
		var parts []string
		for _, e := range x {
			parts = append(parts, valueToString(e))
		}
		return strings.Join(parts, "; ")
	case map[string]interface{}:
		if url, ok := x["value"].(string); ok {
			return url
		}
		return ""
	default:
		return ""
	}
}
