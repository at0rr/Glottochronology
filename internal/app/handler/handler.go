package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"glottochronology/internal/app/middleware"
	"glottochronology/internal/app/redis"
	"glottochronology/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Redis      *redis.Client
}

func NewHandler(r *repository.Repository, rc *redis.Client) *Handler {
	return &Handler{
		Repository: r,
		Redis:      rc,
	}
}

// RegisterHandler — регистрация всех API-маршрутов
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// --- HTML СТРАНИЦЫ (для браузера) ---
	router.GET("/swadesh-lists", h.FeedHandler)             // Сама лента
	router.GET("/swadesh-lists/catalog", h.GridHandler)     // Плитка
	router.GET("/swadesh-lists/new", h.AddHandler)          // Добавление
	router.GET("/swadesh-lists/publish", h.PublishFormHandler) // Публикация

	// --- ПУБЛИЧНЫЕ API маршруты (без авторизации) ---
	api := router.Group("/api")
	api.Use(middleware.OptionalJWTAuth(h.Redis))
	{
		api.POST("/auth/register", h.APIRegister)
		api.POST("/auth/login", h.APILogin)

		api.GET("/swadesh-lists", h.APIGetLanguages)
		api.GET("/swadesh-lists/feed", h.APIGetFeed)
	}

	// --- ЗАЩИЩЁННЫЕ API маршруты (требуют JWT) ---
	auth := router.Group("/api")
	auth.Use(middleware.JWTAuth(h.Redis))
	{
		auth.GET("/swadesh-lists/draft", h.APIGetDraft)
		auth.POST("/swadesh-lists", h.APICreateLanguage)
		auth.PUT("/swadesh-lists/:id/publish", h.APIPublishLanguage)
		auth.DELETE("/swadesh-lists/:id", h.APIDeleteLanguage)
		auth.POST("/swadesh-lists/:id/like", h.APILikeLanguage)
		auth.POST("/auth/logout", h.APILogout)
	}

	// Редирект с корня на HTML-ленту
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swadesh-lists")
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
