package ds

import "glottochronology/internal/app/role"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Login     string    `gorm:"type:varchar(50);unique;not null" json:"login"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Role      role.Role `gorm:"type:int;default:0" json:"role"`
	CreatedAt int64     `gorm:"autoCreateTime" json:"created_at"`
}

func (User) TableName() string {
	return "users"
}
