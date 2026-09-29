package repository

import (
	"gorm.io/gorm"

	"glottochronology/internal/app/ds"
)

// SetLike — action = 1 (поставить), action = 0 (снять)
func (r *Repository) SetLike(userID, languageID uint, action int) error {
	if action == 1 {
		// Проверим, нет ли уже такого лайка
		var count int64
		r.db.Model(&ds.Like{}).
			Where("user_id = ? AND language_id = ?", userID, languageID).
			Count(&count)
		if count > 0 {
			return nil // уже лайкнуто
		}
		return r.db.Create(&ds.Like{UserID: userID, LanguageID: languageID}).Error
	}
	// action == 0 → удаляем
	return r.db.Where("user_id = ? AND language_id = ?", userID, languageID).
		Delete(&ds.Like{}).Error
}

// HasLike — есть ли лайк у пользователя
func (r *Repository) HasLike(userID, languageID uint) bool {
	var count int64
	r.db.Model(&ds.Like{}).
		Where("user_id = ? AND language_id = ?", userID, languageID).
		Count(&count)
	return count > 0
}

var _ = gorm.ErrRecordNotFound
