package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"glottochronology/internal/app/handler"
	"glottochronology/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	// Инициализация репозитория
	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("Ошибка инициализации репозитория")
	}

	// Инициализация хендлера
	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./resources")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// Используем хендлер
	r.GET("/hello", h.GetLanguages)

	r.GET("/language/:id", h.GetLanguage)

	r.Run()
	log.Println("Server down")
}
