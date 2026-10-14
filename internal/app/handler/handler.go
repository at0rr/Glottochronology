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
	// GET
	router.GET("/swadesh-lists", h.FeedHandler)
	router.GET("/swadesh-lists/new", h.AddHandler)
	router.GET("/swadesh-lists/publish", h.PublishFormHandler)
	router.GET("/swadesh-lists/catalog", h.GridHandler)

	// POST
	router.POST("/swadesh-lists/new", h.CreateDraftHandler)
	router.POST("/swadesh-lists/publish", h.PublishHandler)
	router.POST("/swadesh-lists/delete", h.DeleteHandler)

	// редирект с корня
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swadesh-lists")
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
