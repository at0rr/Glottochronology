package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"glottochronology/internal/app/repository"
)

const minioBaseURL = "http://localhost:9000/glotto/"

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func (h *Handler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		published := h.Repository.GetAllPublished()
		if len(published) == 0 {
			c.String(http.StatusNotFound, "Нет доступных услуг")
			return
		}
		h.renderFeed(c, published[0])
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	if c.Query("next") == "true" {
		nextService, err := h.Repository.GetNextAfterID(id)
		if err != nil {
			service, _ := h.Repository.GetByID(id)
			h.renderFeed(c, service)
			return
		}
		h.renderFeed(c, nextService)
		return
	}

	service, err := h.Repository.GetByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Услуга не найдена")
		return
	}
	h.renderFeed(c, service)
}

func (h *Handler) renderFeed(c *gin.Context, service repository.Service) {
	c.HTML(http.StatusOK, "feed.html", gin.H{
		"service":  service,
		"likes":    len(service.Likes),
		"imageURL": minioBaseURL + service.ImageKey,
		"videoURL": minioBaseURL + service.VideoKey,
		"dateStr":  formatYear(service.Date),
	})
}

func (h *Handler) AddHandler(c *gin.Context) {
	draft, err := h.Repository.GetDraft()
	if err != nil {
		c.String(http.StatusNotFound, "Черновик не найден")
		return
	}
	c.HTML(http.StatusOK, "add.html", gin.H{
		"service":  draft,
		"likes":    len(draft.Likes),
		"imageURL": minioBaseURL + draft.ImageKey,
		"videoURL": minioBaseURL + draft.VideoKey,
		"dateStr":  formatYear(draft.Date),
	})
}

func (h *Handler) GridHandler(c *gin.Context) {
	dateFilter := c.Query("date")
	services := h.Repository.GetPublishedFilteredByDate(dateFilter)

	type card struct {
		repository.Service
		ImageURL string
		DateStr  string
		Likes    int
	}
	var cards []card
	for _, s := range services {
		cards = append(cards, card{
			Service:  s,
			ImageURL: minioBaseURL + s.ImageKey,
			DateStr:  formatYear(s.Date),
			Likes:    len(s.Likes),
		})
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"cards":  cards,
		"filter": dateFilter,
	})
}

// formatYear возвращает строку вида "1500 г." или "2000 г. до н.э."
func formatYear(t time.Time) string {
	year := t.Year()
	if year < 0 {
		return fmt.Sprintf("%d г. до н.э.", -year)
	}
	return fmt.Sprintf("%d г.", year)
}
