package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"glottochronology/internal/app/ds"
)

// GetAllPublished — все опубликованные услуги
func (r *Repository) GetAllPublished() ([]ds.Language, error) {
	var langs []ds.Language
	err := r.db.Where("status = ?", "published").Order("id").Find(&langs).Error
	return langs, err
}

// GetByID — одна услуга по ID (кроме deleted)
func (r *Repository) GetByID(id int) (*ds.Language, error) {
	query := `SELECT id, title, description, status, image_key, video_key,
	                 match_percent, divergence_date, date_create, date_formation, creator_id
	          FROM languages WHERE id = $1 AND status != 'deleted'`

	row := r.db.Raw(query, id).Row()

	lang := &ds.Language{}
	err := row.Scan(
		&lang.ID, &lang.Title, &lang.Description, &lang.Status,
		&lang.ImageKey, &lang.VideoKey, &lang.MatchPercent,
		&lang.DivergenceDate, &lang.DateCreate, &lang.DateFormation, &lang.CreatorID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return lang, nil
}

// GetNextAfterID — следующая опубликованная после указанного ID
func (r *Repository) GetNextAfterID(id int) (*ds.Language, error) {
	var lang ds.Language
	err := r.db.Where("status = ? AND id > ?", "published", id).
		Order("id").First(&lang).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lang, nil
}

// GetDraftByCreator — черновик конкретного пользователя
func (r *Repository) GetDraftByCreator(creatorID uint) (*ds.Language, error) {
	var lang ds.Language
	err := r.db.Where("status = ? AND creator_id = ?", "draft", creatorID).
		First(&lang).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lang, nil
}

// GetPublishedFilteredByYear — фильтр по точному году
func (r *Repository) GetPublishedFilteredByYear(year int) ([]ds.Language, error) {
	var langs []ds.Language
	err := r.db.Where("status = ?", "published").
		Where("EXTRACT(YEAR FROM divergence_date) = ?", year).
		Order("id").Find(&langs).Error
	return langs, err
}

// GetPublishedFilteredByDecade — фильтр по десятилетию
func (r *Repository) GetPublishedFilteredByDecade(startYear int) ([]ds.Language, error) {
	var langs []ds.Language
	err := r.db.Where("status = ?", "published").
		Where("EXTRACT(YEAR FROM divergence_date) BETWEEN ? AND ?", startYear, startYear+9).
		Order("id").Find(&langs).Error
	return langs, err
}

// GetPublishedFilteredByCentury — фильтр по столетию
func (r *Repository) GetPublishedFilteredByCentury(startYear int) ([]ds.Language, error) {
	var langs []ds.Language
	err := r.db.Where("status = ?", "published").
		Where("EXTRACT(YEAR FROM divergence_date) BETWEEN ? AND ?", startYear, startYear+99).
		Order("id").Find(&langs).Error
	return langs, err
}

// GetLikeCount — количество лайков для услуги
func (r *Repository) GetLikeCount(languageID uint) int64 {
	var count int64
	r.db.Model(&ds.Like{}).Where("language_id = ?", languageID).Count(&count)
	return count
}

// CreateLanguage — добавление новой услуги-черновика через ORM
func (r *Repository) CreateLanguage(lang *ds.Language) error {
	return r.db.Create(lang).Error
}

// PublishLanguage — публикация (меняет статус) через ORM
func (r *Repository) PublishLanguage(id uint, title, description string, matchPercent int, divergenceDate string) error {
	updates := map[string]interface{}{
		"status":          "published",
		"title":           title,
		"description":     description,
		"match_percent":   matchPercent,
		"divergence_date": divergenceDate,
		"date_formation":  gorm.Expr("now()"),
	}
	return r.db.Model(&ds.Language{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteLanguage — логическое удаление ЧЕРЕЗ SQL UPDATE (без ORM)
func (r *Repository) DeleteLanguage(id uint) error {
	result := r.db.Exec(
		"UPDATE languages SET status = 'deleted' WHERE id = $1 AND status != 'deleted'",
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
