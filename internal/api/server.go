package api

import (
	"log"

	"github.com/gin-gonic/gin"
	"glottochronology/internal/app/handler"
	"glottochronology/internal/app/repository"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		log.Println("Ошибка инициализации репозитория:", err)
		return
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/feed", h.FeedHandler)
	r.GET("/add", h.AddHandler)
	r.GET("/grid", h.GridHandler)

	// Редирект с корня на ленту
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/feed")
	})

	if err := r.Run(); err != nil {
		log.Println("Ошибка запуска сервера:", err)
	}
	log.Println("Server down")
}
