package mysql

import (
	"github.com/formsx/backend/internal/models"
	"gorm.io/gorm"
)

type FormRepo struct{ db *gorm.DB }

func NewFormRepo(db *gorm.DB) *FormRepo { return &FormRepo{db: db} }

func (r *FormRepo) Create(f *models.Form) error {
	return r.db.Create(f).Error
}

func (r *FormRepo) GetByID(id int64) (*models.Form, error) {
	var f models.Form
	err := r.db.First(&f, id).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FormRepo) GetBySlug(slug string) (*models.Form, error) {
	var f models.Form
	err := r.db.Where("slug = ?", slug).First(&f).Error
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FormRepo) List(page, limit int, search string) ([]models.Form, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	q := r.db.Model(&models.Form{})
	if search != "" {
		q = q.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Form
	offset := (page - 1) * limit
	if err := q.Offset(offset).Limit(limit).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *FormRepo) Update(f *models.Form) error {
	return r.db.Save(f).Error
}

func (r *FormRepo) Delete(id int64) error {
	return r.db.Delete(&models.Form{}, id).Error
}

func (r *FormRepo) ExistsSlug(slug string, excludeID int64) (bool, error) {
	var n int64
	q := r.db.Model(&models.Form{}).Where("slug = ?", slug)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&n).Error
	return n > 0, err
}
