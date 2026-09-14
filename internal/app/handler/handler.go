package handler

import (
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

// RegisterHandler — все маршруты
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// 3 GET
	router.GET("/feed", h.FeedHandler)
	router.GET("/add", h.AddHandler)
	router.GET("/grid", h.GridHandler)

	// 3 POST
	router.POST("/add", h.CreateDraftHandler)
	router.POST("/publish", h.PublishHandler)
	router.POST("/delete", h.DeleteHandler)

	// редирект с корня
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/feed")
	})
}

// RegisterStatic — статика и шаблоны
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, code int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(code, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
