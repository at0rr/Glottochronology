package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"glottochronology/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

// RegisterHandler — регистрация всех API-маршрутов
func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Услуги
		api.GET("/swadesh-lists", h.APIGetLanguages)
		api.GET("/swadesh-lists/feed", h.APIGetFeed)
		api.GET("/swadesh-lists/draft", h.APIGetDraft)
		api.POST("/swadesh-lists", h.APICreateLanguage)
		api.PUT("/swadesh-lists/:id/publish", h.APIPublishLanguage)
		api.DELETE("/swadesh-lists/:id", h.APIDeleteLanguage)
		api.POST("/swadesh-lists/:id/like", h.APILikeLanguage)

		// Пользователи
		api.POST("/auth/register", h.APIRegister)
		api.POST("/auth/login", h.APILogin)
		api.POST("/auth/logout", h.APILogout)
	}

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/api/swadesh-lists")
	})
}

// RegisterStatic — статика (картинки-заглушки)
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, code int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(code, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
