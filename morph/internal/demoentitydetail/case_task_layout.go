package demoentitydetail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CaseTaskDemoDetailJSON builds a realistic detail JSON payload for case/task entity_details.
func CaseTaskDemoDetailJSON(id int, title, assigneeType string, assigneeID int, startAt, endAt *time.Time) string {
	h := hashDemoDetail(id ^ 0x4354534b) // "CTSK"

	priorities := []string{"low", "normal", "high", "urgent"}
	statuses := []string{"new", "in_progress", "blocked", "pending_review"}
	categories := []string{"operations", "member_support", "fleet", "routing", "compliance"}
	riskNotes := []string{
		"No immediate rider impact expected.",
		"Potential delay impact during peak AM window.",
		"Needs cross-team review before closure.",
		"Awaiting external confirmation document.",
	}
	checklist := []string{
		"Validate assignment and ownership",
		"Review timeline with dispatcher",
		"Confirm supporting documents uploaded",
		"Post final note and closure summary",
	}

	startText := ""
	endText := ""
	if startAt != nil {
		startText = startAt.UTC().Format(time.RFC3339)
	}
	if endAt != nil {
		endText = endAt.UTC().Format(time.RFC3339)
	}

	typ := strings.ToLower(strings.TrimSpace(assigneeType))
	if typ == "" {
		typ = "employee"
	}

	obj := map[string]interface{}{
		"case_summary": map[string]interface{}{
			"title":                 strings.TrimSpace(title),
			"demo_seed_record_id":   id,
			"priority":              priorities[int(h%uint64(len(priorities)))],
			"status":                statuses[int((h>>8)%uint64(len(statuses)))],
			"category":              categories[int((h>>16)%uint64(len(categories)))],
			"target_start_utc":      startText,
			"target_end_utc":        endText,
			"estimated_minutes":     30 + int((h>>20)%180),
			"assignment_owner_type": typ,
			"assignment_owner_id":   assigneeID,
		},
		"workflow": map[string]interface{}{
			"next_action":           checklist[int((h>>24)%uint64(len(checklist)))],
			"review_required":       (h % 3) == 0,
			"document_check_needed": (h % 2) == 0,
			"risk_note":             riskNotes[int((h>>28)%uint64(len(riskNotes)))],
		},
		"tags": []string{
			fmt.Sprintf("seed-%d", id%50),
			categories[int((h>>16)%uint64(len(categories)))],
			statuses[int((h>>8)%uint64(len(statuses)))],
		},
		"audit": map[string]interface{}{
			"created_by_seed": "cmd/seed_tran",
			"generated_at":    time.Now().UTC().Format(time.RFC3339),
		},
	}

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
