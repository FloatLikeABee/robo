package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	bkAssistantIDPrefix = "bk:"
	bkRAGMaxCollections = 4
	bkRAGResultsPerCol  = 4
	bkRAGSnippetRunes   = 900
	bkRAGTotalRunes     = 4500
)

type bkAssistantProfile struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	SystemPrompt   string   `json:"system_prompt"`
	RAGCollections []string `json:"rag_collections"`
	LLMProvider    string   `json:"llm_provider"`
	ModelName      string   `json:"model_name"`
}

func morphAgentIDForBK(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if strings.HasPrefix(id, bkAssistantIDPrefix) {
		return id
	}
	return bkAssistantIDPrefix + id
}

func parseBKAssistantID(agentID string) (string, bool) {
	agentID = strings.TrimSpace(agentID)
	if !strings.HasPrefix(agentID, bkAssistantIDPrefix) {
		return "", false
	}
	id := strings.TrimSpace(strings.TrimPrefix(agentID, bkAssistantIDPrefix))
	if id == "" {
		return "", false
	}
	return id, true
}

func (h *Handlers) bkAPIBase() string {
	return strings.TrimSuffix(strings.TrimSpace(h.externalAPIBase), "/")
}

func (h *Handlers) bkHTTPGetJSON(ctx context.Context, path string, dest any) error {
	base := h.bkAPIBase()
	if base == "" {
		return fmt.Errorf("EXTERNAL_API_BASE not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("bk GET %s: HTTP %d: %s", path, res.StatusCode, truncateRunes(string(body), 180))
	}
	return json.Unmarshal(body, dest)
}

func (h *Handlers) bkHTTPPostJSON(ctx context.Context, path string, payload any, dest any) error {
	base := h.bkAPIBase()
	if base == "" {
		return fmt.Errorf("EXTERNAL_API_BASE not configured")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 45 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("bk POST %s: HTTP %d: %s", path, res.StatusCode, truncateRunes(string(body), 180))
	}
	if dest == nil {
		return nil
	}
	return json.Unmarshal(body, dest)
}

func (h *Handlers) listBKAssistants(ctx context.Context) ([]bkAssistantProfile, error) {
	var list []bkAssistantProfile
	if err := h.bkHTTPGetJSON(ctx, "/assistants", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (h *Handlers) getBKAssistant(ctx context.Context, id string) (*bkAssistantProfile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("missing assistant id")
	}
	var out bkAssistantProfile
	path := "/assistants/" + url.PathEscape(id)
	if err := h.bkHTTPGetJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	if strings.TrimSpace(out.ID) == "" {
		out.ID = id
	}
	return &out, nil
}

func (h *Handlers) queryBKRAGCollection(ctx context.Context, collection, query string, n int) ([]string, error) {
	collection = strings.TrimSpace(collection)
	query = strings.TrimSpace(query)
	if collection == "" || query == "" {
		return nil, nil
	}
	if n <= 0 {
		n = bkRAGResultsPerCol
	}
	var resp struct {
		Results []struct {
			Content string `json:"content"`
		} `json:"results"`
	}
	path := "/rag/collections/" + url.PathEscape(collection) + "/query"
	if err := h.bkHTTPPostJSON(ctx, path, map[string]any{
		"query":     query,
		"n_results": n,
	}, &resp); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(resp.Results))
	for _, r := range resp.Results {
		c := strings.TrimSpace(r.Content)
		if c != "" {
			out = append(out, c)
		}
	}
	return out, nil
}

func truncateRunesSoft(s string, max int) string {
	s = strings.TrimSpace(s)
	if s == "" || max <= 0 {
		return s
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max]) + "…"
}

// buildBKAssistantInstructions loads a BK assistant and optional RAG snippets for Morph chat.
func (h *Handlers) buildBKAssistantInstructions(ctx context.Context, bkID, userMessage string) (name, instructions string, err error) {
	profile, err := h.getBKAssistant(ctx, bkID)
	if err != nil {
		return "", "", err
	}
	var b strings.Builder
	b.WriteString("You are running as AI tools assistant \"" + strings.TrimSpace(profile.Name) + "\".\n")
	prompt := strings.TrimSpace(profile.SystemPrompt)
	if prompt != "" {
		b.WriteString("\n--- Assistant system prompt ---\n")
		b.WriteString(prompt)
		b.WriteString("\n")
	}
	cols := profile.RAGCollections
	if len(cols) > bkRAGMaxCollections {
		cols = cols[:bkRAGMaxCollections]
	}
	if len(cols) > 0 && strings.TrimSpace(userMessage) != "" {
		var ragParts []string
		total := 0
		for _, col := range cols {
			snips, qerr := h.queryBKRAGCollection(ctx, col, userMessage, bkRAGResultsPerCol)
			if qerr != nil {
				log.Printf("[BK-ASSISTANT] RAG query %q: %v", col, qerr)
				continue
			}
			for i, snip := range snips {
				snip = truncateRunesSoft(snip, bkRAGSnippetRunes)
				if snip == "" {
					continue
				}
				block := fmt.Sprintf("[%s #%d]\n%s", col, i+1, snip)
				if total+utf8.RuneCountInString(block) > bkRAGTotalRunes {
					break
				}
				ragParts = append(ragParts, block)
				total += utf8.RuneCountInString(block)
			}
			if total >= bkRAGTotalRunes {
				break
			}
		}
		if len(ragParts) > 0 {
			b.WriteString("\n--- Retrieved RAG context (use when relevant; cite collection names lightly) ---\n")
			b.WriteString(strings.Join(ragParts, "\n\n"))
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(profile.Name), strings.TrimSpace(b.String()), nil
}
