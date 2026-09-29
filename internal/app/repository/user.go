package repository

import (
	"errors"

	"gorm.io/gorm"

	"glottochronology/internal/app/ds"
)

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

func (r *Repository) GetUserByID(id uint) (*ds.User, error) {
	var u ds.User
	err := r.db.Where("id = ?", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

func (r *Repository) CreateUser(u *ds.User) error {
	return r.db.Create(u).Error
}
