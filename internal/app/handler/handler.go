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
	// --- ПУБЛИЧНЫЕ маршруты (без авторизации) ---
	api := router.Group("/api")
	{
		// Auth
		api.POST("/auth/register", h.APIRegister)
		api.POST("/auth/login", h.APILogin)

		// Чтение (GET) — доступно всем
		api.GET("/languages", h.APIGetLanguages)
		api.GET("/languages/feed", h.APIGetFeed)
		api.GET("/languages/:id", h.APIGetFeed)
	}

	// --- ЗАЩИЩЁННЫЕ маршруты (требуют JWT) ---
	auth := router.Group("/api")
	auth.Use(middleware.JWTAuth(h.Redis))
	{
		// Черновик — только для авторизованного
		auth.GET("/languages/draft", h.APIGetDraft)

		// Методы создателя
		auth.POST("/languages", h.APICreateLanguage)
		auth.PUT("/languages/:id/publish", h.APIPublishLanguage)
		auth.DELETE("/languages/:id", h.APIDeleteLanguage)
		auth.POST("/languages/:id/like", h.APILikeLanguage)

		// Logout — только для авторизованного
		auth.POST("/auth/logout", h.APILogout)
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
