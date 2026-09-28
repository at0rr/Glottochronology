package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"glottochronology/internal/app/ds"
)

// GetAllPublished — все опубликованные услуги (кроме deleted)
func (r *Repository) GetAllPublished() ([]ds.Language, error) {
	var langs []ds.Language
	err := r.db.Where("status = ?", "published").Order("id").Find(&langs).Error
	if err != nil {
		return nil, err
	}
	return langs, nil
}

// GetByID — одна услуга по ID (любого статуса кроме deleted)
func (r *Repository) GetByID(id int) (*ds.Language, error) {
	query := `SELECT id, title, description, status, image_url, video_url,
	                 match_percent, divergence_date, date_create, date_formation, creator_id
	          FROM languages WHERE id = $1 AND status != 'deleted'`

	row := r.db.Raw(query, id).Row()

	lang := &ds.Language{}
	err := row.Scan(
		&lang.ID, &lang.Title, &lang.Description, &lang.Status,
		&lang.ImageURL, &lang.VideoURL, &lang.MatchPercent,
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

// GetPublishedFilteredByDate — фильтр по году
func (r *Repository) GetPublishedFilteredByDate(dateFilter string) ([]ds.Language, error) {
	var langs []ds.Language
	q := r.db.Where("status = ?", "published")
	if dateFilter != "" {
		q = q.Where("EXTRACT(YEAR FROM divergence_date) = ?", dateFilter)
	}
	err := q.Order("id").Find(&langs).Error
	return langs, err
}

// GetLikeCount — количество лайков для услуги (для отображения)
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

// DeleteLanguage — логическое удаление ЧЕРЕЗ SQL UPDATE (без ORM), как требует задание
func (r *Repository) DeleteLanguage(id uint) error {
	result := r.db.Exec("UPDATE languages SET status = 'deleted' WHERE id = $1 AND status != 'deleted'", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("услуга с id %d не найдена или уже удалена", id)
	}
	return nil
}

// GetUniqueDivergenceYears возвращает уникальные годы расхождения из БД
func (r *Repository) GetUniqueDivergenceYears() ([]int, error) {
	var years []int
	err := r.db.Model(&ds.Language{}).
		Where("status = ?", "published").
		Distinct().
		Pluck("EXTRACT(YEAR FROM divergence_date)::int", &years).Error
	if err != nil {
		return nil, err
	}
	return years, nil
}
