package mysql

import (
	"github.com/formsx/backend/internal/models"
	"gorm.io/gorm"
)

type QuestionRepo struct{ db *gorm.DB }

func NewQuestionRepo(db *gorm.DB) *QuestionRepo { return &QuestionRepo{db: db} }

func (r *QuestionRepo) Create(q *models.Question) error {
	return r.db.Create(q).Error
}

func (r *QuestionRepo) GetByID(id int64) (*models.Question, error) {
	var q models.Question
	err := r.db.First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *QuestionRepo) ListByFormID(formID int64) ([]models.Question, error) {
	var list []models.Question
	err := r.db.Model(&models.Question{}).
		Joins("LEFT JOIN form_pages ON form_pages.id = questions.page_id AND form_pages.deleted_at IS NULL").
		Where("questions.form_id = ?", formID).
		Order("form_pages.sort_order ASC, questions.sort_order ASC, questions.id ASC").
		Find(&list).Error
	return list, err
}

func (r *QuestionRepo) ReassignPage(formID, fromPageID, toPageID int64) error {
	return r.db.Model(&models.Question{}).
		Where("form_id = ? AND page_id = ?", formID, fromPageID).
		Update("page_id", toPageID).Error
}

func (r *QuestionRepo) Update(q *models.Question) error {
	return r.db.Save(q).Error
}

func (r *QuestionRepo) Delete(id int64) error {
	return r.db.Delete(&models.Question{}, id).Error
}

func (r *QuestionRepo) CountByFormID(formID int64) (int64, error) {
	var n int64
	err := r.db.Model(&models.Question{}).Where("form_id = ?", formID).Count(&n).Error
	return n, err
}

func (r *QuestionRepo) GetByFormIDAndID(formID, questionID int64) (*models.Question, error) {
	var q models.Question
	err := r.db.Where("form_id = ? AND id = ?", formID, questionID).First(&q).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}
