package ds

// Like — связь пользователя и услуги (м-м). Уникальный ключ по паре.
type Like struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint `gorm:"not null;uniqueIndex:idx_user_lang"`
	LanguageID uint `gorm:"not null;uniqueIndex:idx_user_lang"`

	User     User     `gorm:"foreignKey:UserID"`
	Language Language `gorm:"foreignKey:LanguageID"`
}
