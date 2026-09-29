package ds

type User struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Login     string `gorm:"type:varchar(50);unique;not null" json:"login"`
	Password  string `gorm:"type:varchar(100);not null" json:"-"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
