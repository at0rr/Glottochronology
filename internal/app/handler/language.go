package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"glottochronology/internal/app/ds"
)

// currentUserID — временно захардкоженный ID пользователя.
// В следующих лабах замените на получение из JWT.
const currentUserID uint = 1

const defaultImage = "/static/img/default.jpg"
const defaultVideo = "/static/video/default.mp4"

// imageOr — возвращает URL или путь к дефолтной картинке
func imageOr(url string) string {
	if url == "" {
		return defaultImage
	}
	return url
}

func videoOr(url string) string {
	if url == "" {
		return defaultVideo
	}
	return url
}

func formatYear(t time.Time) string {
	y := t.Year()
	if y < 0 {
		return strconv.Itoa(-y) + " г. до н.э."
	}
	return strconv.Itoa(y) + " г."
}

// --- GET /feed ---
func (h *Handler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")

	var lang *ds.Language
	var err error

	if idStr == "" {
		// если id не указан — берём первую опубликованную
		all, e := h.Repository.GetAllPublished()
		if e != nil || len(all) == 0 {
			c.String(http.StatusNotFound, "Нет услуг")
			return
		}
		lang = &all[0]
	} else {
		id, e := strconv.Atoi(idStr)
		if e != nil {
			c.String(http.StatusBadRequest, "Неверный ID")
			return
		}
		if c.Query("next") == "true" {
			next, e := h.Repository.GetNextAfterID(id)
			if e != nil || next == nil {
				lang, _ = h.Repository.GetByID(id)
			} else {
				lang = next
			}
		} else {
			lang, err = h.Repository.GetByID(id)
			if err != nil || lang == nil {
				c.String(http.StatusNotFound, "Услуга не найдена")
				return
			}
		}
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"service":  lang,
		"likes":    h.Repository.GetLikeCount(lang.ID),
		"imageURL": imageOr(lang.ImageURL),
		"videoURL": videoOr(lang.VideoURL),
		"dateStr":  formatYear(lang.DivergenceDate),
	})
}

// --- GET /add ---
func (h *Handler) AddHandler(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// если черновика нет — отдаём пустую структуру
	if draft == nil {
		draft = &ds.Language{}
	}

	c.HTML(http.StatusOK, "add.html", gin.H{
		"service":  draft,
		"likes":    0,
		"imageURL": imageOr(draft.ImageURL),
		"videoURL": videoOr(draft.VideoURL),
		"dateStr":  "",
		"hasDraft": draft.ID != 0,
	})
}

// --- GET /grid ---
func (h *Handler) GridHandler(c *gin.Context) {
	dateFilter := c.Query("date")
	langs, err := h.Repository.GetPublishedFilteredByDate(dateFilter)
	if err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	type card struct {
		ds.Language
		ImageURL string
		DateStr  string
		Likes    int64
	}

	var cards []card
	for _, l := range langs {
		cards = append(cards, card{
			Language: l,
			ImageURL: imageOr(l.ImageURL),
			DateStr:  formatYear(l.DivergenceDate),
			Likes:    h.Repository.GetLikeCount(l.ID),
		})
	}

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"cards":  cards,
		"filter": dateFilter,
	})
}

// --- POST /add ---
// Создание новой услуги-черновика. Из формы берём только название.
// Фото и видео в этой лабе НЕ сохраняются и НЕ передаются на сервер.
func (h *Handler) CreateDraftHandler(c *gin.Context) {
    title := c.PostForm("title")  // берём только название

    lang := &ds.Language{
        Title:          title,
        Description:    "",
        Status:         "draft",
        ImageURL:       "",
        VideoURL:       "",
        MatchPercent:   0,
        DivergenceDate: time.Now(),
        CreatorID:      currentUserID,
    }

    if err := h.Repository.CreateLanguage(lang); err != nil {
        h.errorHandler(c, http.StatusInternalServerError, err)
        return
    }

    c.Redirect(http.StatusFound, "/add")
}

// --- POST /publish ---
// Публикация: заполняем описание, оба поля по теме.
func (h *Handler) PublishHandler(c *gin.Context) {
    idStr := c.PostForm("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.String(http.StatusBadRequest, "Неверный ID")
        return
    }

    title := c.PostForm("title")
    description := c.PostForm("description")
    matchStr := c.PostForm("match_percent")
    divDate := c.PostForm("divergence_date")

    match, _ := strconv.Atoi(matchStr)

    if err := h.Repository.PublishLanguage(uint(id), title, description, match, divDate); err != nil {
        h.errorHandler(c, http.StatusInternalServerError, err)
        return
    }

    c.Redirect(http.StatusFound, "/grid")
}

// --- POST /delete ---
// Логическое удаление через SQL UPDATE (без ORM)
func (h *Handler) DeleteHandler(c *gin.Context) {
	idStr := c.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	if err := h.Repository.DeleteLanguage(uint(id)); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/grid")
}
