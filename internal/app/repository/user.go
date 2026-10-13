package repository

import (
	"errors"

	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"

	"glottochronology/internal/app/ds"
	"glottochronology/internal/app/role"
)

// Register — регистрация нового пользователя (пароль хешируется)
func (r *Repository) Register(login, password string) (*ds.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &ds.User{
		Login:    login,
		Password: string(hash),
		Role:     role.User,
	}

	if err := r.db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByLogin — поиск по логину
func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var u ds.User
	err := r.db.Where("login = ?", login).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByID — поиск по ID
func (r *Repository) GetUserByID(id uint) (*ds.User, error) {
	var u ds.User
	err := r.db.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// CheckPassword — сравнение хеша пароля
func (r *Repository) CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
