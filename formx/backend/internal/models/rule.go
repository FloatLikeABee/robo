package models

import (
	"time"

	"gorm.io/gorm"
)

// QuestionRule controls visibility of a question based on another question's answer.
// Rule applies TO question_id (show/hide this question); depends_on_question_id is the one we check.
// Condition: "answered" = show this question only when depends_on is answered; "not_answered" = show when not answered.
// Multiple rules on the same question are ANDed (all must pass). One rule per (question_id, depends_on_question_id) so no negation.
const (
	RuleConditionAnswered    = "answered"
	RuleConditionNotAnswered = "not_answered"
)

var ValidRuleConditions = map[string]bool{
	RuleConditionAnswered: true, RuleConditionNotAnswered: true,
}

type QuestionRule struct {
	ID                  int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	QuestionID          int64          `json:"question_id" gorm:"uniqueIndex:idx_rule_question_depends;not null"` // question this rule applies to
	DependsOnQuestionID int64          `json:"depends_on_question_id" gorm:"uniqueIndex:idx_rule_question_depends;not null"`
	Condition           string         `json:"condition" gorm:"size:32;not null"` // "answered" | "not_answered"
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

func (QuestionRule) TableName() string { return "question_rules" }

type CreateRuleRequest struct {
	DependsOnQuestionID int64  `json:"depends_on_question_id" binding:"required"`
	Condition           string `json:"condition" binding:"required"`
}
