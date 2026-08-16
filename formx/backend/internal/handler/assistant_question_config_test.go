package handler

import (
	"testing"

	"github.com/formsx/backend/internal/models"
)

func TestParseQuestionConfigFromAI_selectShorthand(t *testing.T) {
	raw := map[string]interface{}{
		"options": []interface{}{"Yes", "No", "Maybe"},
	}
	cfg, err := parseQuestionConfigFromAI(raw, models.QTypeSelect)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(cfg.Options))
	}
	if cfg.Options[0].Label != "Yes" || cfg.Options[0].Value != 1 {
		t.Fatalf("unexpected first option: %+v", cfg.Options[0])
	}
}

func TestParseQuestionConfigFromAI_selectRequiresOptions(t *testing.T) {
	_, err := parseQuestionConfigFromAI(nil, models.QTypeSelect)
	if err == nil {
		t.Fatal("expected error for select without options")
	}
}

func TestParseQuestionConfigFromAI_choicesAlias(t *testing.T) {
	raw := map[string]interface{}{
		"choices": []interface{}{
			map[string]interface{}{"value": 10, "label": "A"},
			map[string]interface{}{"value": 20, "label": "B"},
		},
	}
	cfg, err := parseQuestionConfigFromAI(raw, models.QTypeMultiselect)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Options) != 2 || cfg.Options[1].Value != 20 {
		t.Fatalf("unexpected options: %+v", cfg.Options)
	}
}
