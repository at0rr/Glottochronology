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

	// 1. Добавляем новые модели Attribute и AttributeValue
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Language{},
		&ds.Like{},
		&ds.Attribute{},      // Таблица справочника атрибутов
		&ds.AttributeValue{}, // Связующая таблица значений
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

	// 2. Добавляем пару базовых атрибутов для тестов (если их еще нет)
	var attrCount int64
	db.Model(&ds.Attribute{}).Count(&attrCount)
	if attrCount == 0 {
		attrs := []ds.Attribute{
			{Name: "Языковая семья"},
			{Name: "Регион"},
			{Name: "Количество носителей"},
		}
		if err := db.Create(&attrs).Error; err != nil {
			panic("error creating default attributes: " + err.Error())
		}
	}

	println("Migration complete")
}
