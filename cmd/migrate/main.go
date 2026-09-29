package main

import (
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"glottochronology/internal/app/ds"
	"glottochronology/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.Language{},
		&ds.Like{},
	)
	if err != nil {
		panic("cant migrate db: " + err.Error())
	}

	// Создаём дефолтного пользователя id=1, если его нет
	var count int64
	db.Model(&ds.User{}).Where("login = ?", "default").Count(&count)
	if count == 0 {
		u := ds.User{Login: "default", Password: "default"}
		if err := db.Create(&u).Error; err != nil {
			panic("error creating default user: " + err.Error())
		}
	}

	println("Migration complete")
}
