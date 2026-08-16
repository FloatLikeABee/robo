package mysql

import (
	"github.com/formsx/backend/internal/models"
	"gorm.io/gorm"
)

type RuleRepo struct{ db *gorm.DB }

func NewRuleRepo(db *gorm.DB) *RuleRepo { return &RuleRepo{db: db} }

func (r *RuleRepo) Create(rule *models.QuestionRule) error {
	return r.db.Create(rule).Error
}

func (r *RuleRepo) GetByID(id int64) (*models.QuestionRule, error) {
	var rule models.QuestionRule
	err := r.db.First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *RuleRepo) ListByQuestionID(questionID int64) ([]models.QuestionRule, error) {
	var list []models.QuestionRule
	err := r.db.Where("question_id = ?", questionID).Find(&list).Error
	return list, err
}

// ListByFormID returns all rules for questions belonging to the form (for public form payload).
func (r *RuleRepo) ListByFormID(formID int64) ([]models.QuestionRule, error) {
	var list []models.QuestionRule
	err := r.db.Joins("JOIN questions ON questions.id = question_rules.question_id").
		Where("questions.form_id = ?", formID).
		Find(&list).Error
	return list, err
}

func (r *RuleRepo) Delete(id int64) error {
	return r.db.Delete(&models.QuestionRule{}, id).Error
}

func (r *RuleRepo) ExistsForQuestionAndDepends(questionID, dependsOnID int64) (bool, error) {
	var n int64
	err := r.db.Model(&models.QuestionRule{}).
		Where("question_id = ? AND depends_on_question_id = ?", questionID, dependsOnID).
		Count(&n).Error
	return n > 0, err
}
