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
		api.GET("/languages", h.APIGetLanguages)
		api.GET("/languages/draft", h.APIGetDraft)
		api.GET("/languages/feed", h.APIGetFeed)
		api.GET("/languages/:id", h.APIGetFeed)
		api.POST("/languages", h.APICreateLanguage)
		api.PUT("/languages/:id/publish", h.APIPublishLanguage)
		api.DELETE("/languages/:id", h.APIDeleteLanguage)
		api.POST("/languages/:id/like", h.APILikeLanguage)

		// Пользователи
		api.POST("/auth/register", h.APIRegister)
		api.POST("/auth/login", h.APILogin)
		api.POST("/auth/logout", h.APILogout)
	}

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/api/languages")
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
