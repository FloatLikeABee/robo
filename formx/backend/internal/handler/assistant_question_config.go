package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/formsx/backend/internal/models"
)

// parseQuestionConfigFromAI normalizes LLM-supplied config (including loose option shapes).
func parseQuestionConfigFromAI(raw interface{}, qType string) (models.QuestionConfig, error) {
	m := map[string]interface{}{}
	if raw != nil {
		switch v := raw.(type) {
		case map[string]interface{}:
			m = v
		default:
			b, _ := json.Marshal(raw)
			_ = json.Unmarshal(b, &m)
		}
	}
	if _, ok := m["options"]; !ok {
		for _, alias := range []string{"choices", "items", "values"} {
			if v, ok := m[alias]; ok {
				m["options"] = v
				break
			}
		}
	}
	if rawOpts, ok := m["options"]; ok {
		opts, err := normalizeOptionItems(rawOpts)
		if err != nil {
			return models.QuestionConfig{}, err
		}
		if len(opts) > 0 {
			m["options"] = opts
		}
	}

	b, err := json.Marshal(m)
	if err != nil {
		return models.QuestionConfig{}, err
	}
	var config models.QuestionConfig
	if err := json.Unmarshal(b, &config); err != nil {
		return models.QuestionConfig{}, err
	}
	if qType == models.QTypeSelect || qType == models.QTypeMultiselect {
		if len(config.Options) == 0 {
			return config, fmt.Errorf(
				"%s requires config.options — e.g. {\"options\":[{\"value\":1,\"label\":\"Yes\"},{\"value\":2,\"label\":\"No\"}]} or [\"Yes\",\"No\"]",
				qType,
			)
		}
	}
	return config, nil
}

func normalizeOptionItems(raw interface{}) ([]models.OptionItem, error) {
	arr, ok := raw.([]interface{})
	if !ok {
		return nil, nil
	}
	out := make([]models.OptionItem, 0, len(arr))
	nextValue := int64(1)
	for _, item := range arr {
		switch v := item.(type) {
		case string:
			label := strings.TrimSpace(v)
			if label == "" {
				continue
			}
			out = append(out, models.OptionItem{Value: nextValue, Label: label})
			nextValue++
		case map[string]interface{}:
			label := firstNonEmptyString(v, "label", "text", "name", "title")
			if label == "" {
				continue
			}
			val := int64(0)
			if n, ok := v["value"].(float64); ok {
				val = int64(n)
			} else if n, ok := v["value"].(int64); ok {
				val = n
			} else if n, ok := v["value"].(int); ok {
				val = int64(n)
			} else if s, ok := v["value"].(string); ok {
				if i, err := parseInt64(strings.TrimSpace(s)); err == nil {
					val = i
				}
			}
			if val == 0 {
				val = nextValue
				nextValue++
			} else if val >= nextValue {
				nextValue = val + 1
			}
			out = append(out, models.OptionItem{Value: val, Label: label})
		default:
			label := strings.TrimSpace(fmt.Sprint(v))
			if label == "" {
				continue
			}
			out = append(out, models.OptionItem{Value: nextValue, Label: label})
			nextValue++
		}
	}
	return out, nil
}

func firstNonEmptyString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok {
			if t := strings.TrimSpace(s); t != "" {
				return t
			}
		}
	}
	return ""
}

func parseInt64(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	var n int64
	_, err := fmt.Sscan(s, &n)
	return n, err
}

func compactQuestionSummary(q models.Question) ginH {
	h := ginH{
		"id":         q.ID,
		"title":      q.Title,
		"type":       q.Type,
		"required":   q.Required,
		"page_id":    q.PageID,
		"sort_order": q.SortOrder,
	}
	if q.Type == models.QTypeSelect || q.Type == models.QTypeMultiselect {
		h["options"] = q.Config.Options
		if q.Config.MaxSelections > 0 {
			h["max_selections"] = q.Config.MaxSelections
		}
	}
	return h
}
