package ds

// User — пользователь системы
type User struct {
	ID        uint   `gorm:"primaryKey"`
	Login     string `gorm:"type:varchar(50);unique;not null"`
	Password  string `gorm:"type:varchar(100);not null"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}
