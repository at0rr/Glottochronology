package handler

import (
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) GetLanguages(c *gin.Context) {
	searchQuery := c.Query("query")
	var languages []repository.Language
	var err error

	if searchQuery == "" {
		languages, err = h.Repository.GetLanguages()
	} else {
		languages, err = h.Repository.SearchLanguagesByTitle(searchQuery)
	}
	if err != nil {
		logrus.Error(err)
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"time":      time.Now().Format("15:04:05"),
		"languages": languages,
		"query":     searchQuery, // сохраняем запрос в поле
	})
}

// GetLanguage обрабатывает GET /language/:id
func (h *Handler) GetLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID:", err)
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	lang, err := h.Repository.GetLanguageByID(id)
	if err != nil {
		logrus.Error(err)
		c.String(http.StatusNotFound, "Язык не найден")
		return
	}

	c.HTML(http.StatusOK, "language.html", gin.H{
		"language": lang,
	})
}
