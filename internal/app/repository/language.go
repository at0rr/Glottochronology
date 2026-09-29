package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"glottochronology/internal/app/ds"
)

// GetAllPublished с фильтром по году/десятилетию/столетию
func (r *Repository) GetAllPublished(mode string, value int, hasFilter bool) ([]ds.Language, error) {
	var langs []ds.Language
	q := r.db.Where("status = ?", "published")

	if hasFilter {
		switch mode {
		case "year":
			q = q.Where("EXTRACT(YEAR FROM divergence_date) = ?", value)
		case "decade":
			q = q.Where("EXTRACT(YEAR FROM divergence_date) BETWEEN ? AND ?", value, value+9)
		case "century":
			q = q.Where("EXTRACT(YEAR FROM divergence_date) BETWEEN ? AND ?", value, value+99)
		}
	}
	err := q.Order("id").Find(&langs).Error
	return langs, err
}

// GetByID — одна услуга по ID (не deleted)
func (r *Repository) GetByID(id int) (*ds.Language, error) {
	var lang ds.Language
	err := r.db.Where("id = ? AND status <> ?", id, "deleted").First(&lang).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lang, nil
}

// GetNextAfterID — следующая опубликованная после указанного ID
func (r *Repository) GetNextAfterID(id int) (*ds.Language, error) {
	var lang ds.Language
	err := r.db.Where("status = ? AND id > ?", "published", id).
		Order("id").First(&lang).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &lang, err
}

// GetDraftByCreator — черновик текущего пользователя
func (r *Repository) GetDraftByCreator(creatorID uint) (*ds.Language, error) {
	var lang ds.Language
	err := r.db.Where("status = ? AND creator_id = ?", "draft", creatorID).
		First(&lang).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &lang, err
}

// GetLikeCount — сколько лайков у услуги
func (r *Repository) GetLikeCount(languageID uint) int64 {
	var count int64
	r.db.Model(&ds.Like{}).Where("language_id = ?", languageID).Count(&count)
	return count
}

// CreateDraft — создание черновика
func (r *Repository) CreateDraft(lang *ds.Language) error {
	return r.db.Create(lang).Error
}

// Publish — публикация (только из draft)
func (r *Repository) Publish(id uint, title, description string, matchPercent int, divergenceDate string) error {
	result := r.db.Model(&ds.Language{}).
		Where("id = ? AND status = ?", id, "draft").
		Updates(map[string]interface{}{
			"status":          "published",
			"title":           title,
			"description":     description,
			"match_percent":   matchPercent,
			"divergence_date": divergenceDate,
			"date_formation":  gorm.Expr("now()"),
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("черновик с id %d не найден или уже опубликован", id)
	}
	return nil
}

// SoftDelete — логическое удаление через SQL UPDATE (по ТЗ — без ORM)
func (r *Repository) SoftDelete(id uint) error {
	result := r.db.Exec(
		"UPDATE languages SET status = 'deleted' WHERE id = $1 AND status <> 'deleted'",
		id,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("услуга с id %d не найдена или уже удалена", id)
	}
	return nil
}
