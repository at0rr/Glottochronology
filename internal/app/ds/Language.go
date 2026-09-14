package ds

import "time"

// Language — услуга (список Сводеша для языка)
type Language struct {
	ID             uint      `gorm:"primaryKey"`
	Title          string    `gorm:"type:varchar(100);not null"`   // наименование
	Description    string    `gorm:"type:varchar(500)"`             // краткое описание
	Status         string    `gorm:"type:varchar(20);not null"`     // draft / published / deleted
	ImageURL       string    `gorm:"type:varchar(255)"`             // url к изображению
	VideoURL       string    `gorm:"type:varchar(255)"`             // url к видео
	MatchPercent   int       `gorm:"not null;default:0"`            // поле по теме №1
	DivergenceDate time.Time `gorm:"not null"`                      // поле по теме №2
	DateCreate     time.Time `gorm:"autoCreateTime"`                // дата создания
	DateFormation  time.Time `gorm:"autoUpdateTime"`                // дата формирования
	CreatorID      uint      `gorm:"not null"`                      // создатель

	Creator User `gorm:"foreignKey:CreatorID"`
}
