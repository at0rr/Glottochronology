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

	// ВАЖНО: каскадное удаление запрещено, поэтому используем обычный AutoMigrate.
	// Никаких OnDelete: CASCADE в моделях нет.
	err = db.AutoMigrate(
		&ds.User{},
		&ds.Language{},
		&ds.Like{},
	)
	if err != nil {
		panic("cant migrate db: " + err.Error())
	}

	println("Migration complete")
}
