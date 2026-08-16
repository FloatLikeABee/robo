package validator

import (
	"fmt"

	"github.com/formsx/backend/internal/models"
)

// ValidateSubmission checks required questions and types for answers against form questions.
func ValidateSubmission(questions []models.Question, answers []models.AnswerInput) error {
	byID := make(map[int64]models.Question)
	for _, q := range questions {
		byID[q.ID] = q
	}
	answered := make(map[int64]bool)
	seen := make(map[int64]bool)
	for _, a := range answers {
		if seen[a.QuestionID] {
			return fmt.Errorf("duplicate answer for question_id %d", a.QuestionID)
		}
		seen[a.QuestionID] = true
		q, ok := byID[a.QuestionID]
		if !ok {
			return fmt.Errorf("unknown question_id %d", a.QuestionID)
		}
		if err := validateAnswerType(q.Type, a.Value); err != nil {
			return fmt.Errorf("question %d (%s): %w", q.ID, q.Title, err)
		}
		answered[q.ID] = true
	}
	for _, q := range questions {
		if q.Required && !answered[q.ID] {
			return fmt.Errorf("required question missing: %s", q.Title)
		}
	}
	return nil
}

func validateAnswerType(qType string, value interface{}) error {
	if value == nil {
		return fmt.Errorf("value is required")
	}
	switch qType {
	case models.QTypeText, models.QTypeQRCode:
		// string
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string")
		}
	case models.QTypeInteger:
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer")
			}
		case int, int64:
		default:
			return fmt.Errorf("expected integer")
		}
	case models.QTypeFloat:
		switch value.(type) {
		case float64, int, int64:
		default:
			return fmt.Errorf("expected number")
		}
	case models.QTypeBoolean:
		switch value.(type) {
		case bool:
		default:
			return fmt.Errorf("expected boolean")
		}
	case models.QTypeSelect:
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer for select")
			}
		case int, int64:
		default:
			return fmt.Errorf("expected integer for select")
		}
	case models.QTypeMultiselect:
		_, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for multiselect")
		}
	case models.QTypeDate, models.QTypeDatetime:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected date string")
		}
	case models.QTypeImage, models.QTypeDocument:
		// value can be string (URL/path) or object with url, filename, size
		switch value.(type) {
		case string:
		case map[string]interface{}:
		default:
			return fmt.Errorf("expected string or file object")
		}
	default:
		return fmt.Errorf("unsupported question type %s", qType)
	}
	return nil
}
