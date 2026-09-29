package ds

type Like struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	UserID     uint `gorm:"not null;uniqueIndex:idx_user_lang" json:"user_id"`
	LanguageID uint `gorm:"not null;uniqueIndex:idx_user_lang" json:"language_id"`

	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Language Language `gorm:"foreignKey:LanguageID" json:"-"`
}

func (Like) TableName() string {
	return "likes"
}
