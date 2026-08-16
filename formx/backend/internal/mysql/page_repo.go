package mysql

import (
	"github.com/formsx/backend/internal/models"
	"gorm.io/gorm"
)

type PageRepo struct{ db *gorm.DB }

func NewPageRepo(db *gorm.DB) *PageRepo { return &PageRepo{db: db} }

func (r *PageRepo) Create(p *models.FormPage) error {
	return r.db.Create(p).Error
}

func (r *PageRepo) ListByFormID(formID int64) ([]models.FormPage, error) {
	var list []models.FormPage
	err := r.db.Where("form_id = ?", formID).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *PageRepo) GetByFormIDAndID(formID, pageID int64) (*models.FormPage, error) {
	var p models.FormPage
	err := r.db.Where("form_id = ? AND id = ?", formID, pageID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PageRepo) Update(p *models.FormPage) error {
	return r.db.Save(p).Error
}

func (r *PageRepo) Delete(id int64) error {
	return r.db.Delete(&models.FormPage{}, id).Error
}

func (r *PageRepo) CountByFormID(formID int64) (int64, error) {
	var n int64
	err := r.db.Model(&models.FormPage{}).Where("form_id = ?", formID).Count(&n).Error
	return n, err
}

// EnsureDefaultPage creates one empty-named page when a form has none, and assigns orphan questions to it.
func (r *PageRepo) EnsureDefaultPage(formID int64) (*models.FormPage, error) {
	list, err := r.ListByFormID(formID)
	if err != nil {
		return nil, err
	}
	if len(list) > 0 {
		page := list[0]
		if err := r.db.Model(&models.Question{}).Where("form_id = ? AND (page_id IS NULL OR page_id = 0)", formID).
			Update("page_id", page.ID).Error; err != nil {
			return nil, err
		}
		return &page, nil
	}
	page := &models.FormPage{FormID: formID, Name: "", SortOrder: 0}
	if err := r.Create(page); err != nil {
		return nil, err
	}
	if err := r.db.Model(&models.Question{}).Where("form_id = ? AND (page_id IS NULL OR page_id = 0)", formID).
		Update("page_id", page.ID).Error; err != nil {
		return nil, err
	}
	return page, nil
}
