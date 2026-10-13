package ds

import "time"

// Language — услуга (список Сводеша для языка)
type Language struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `gorm:"type:varchar(100);not null" json:"title"`
	Description    string    `gorm:"type:varchar(500)" json:"description"`
	Status         string    `gorm:"type:varchar(20);not null" json:"status"`
	ImageKey       string    `gorm:"column:image_key;type:varchar(255)" json:"image_key"`
	VideoKey       string    `gorm:"column:video_key;type:varchar(255)" json:"video_key"`
	MatchPercent   int       `gorm:"not null;default:0" json:"match_percent"`
	DivergenceDate time.Time `gorm:"not null" json:"divergence_date"`
	DateCreate     time.Time `gorm:"autoCreateTime" json:"date_create"`
	DateFormation  time.Time `gorm:"autoUpdateTime" json:"date_formation"`
	CreatorID      uint      `gorm:"not null" json:"creator_id"`
}

func (Language) TableName() string {
	return "languages"
}
