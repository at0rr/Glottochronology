package handler

import (
	"net/http"
	"strconv"
	"time"
	"strings"

	"github.com/gin-gonic/gin"

	"glottochronology/internal/app/ds"
)

// currentUserID — временно захардкоженный ID пользователя.
// В следующих лабах замените на получение из JWT.
const currentUserID uint = 1

// Базовый URL для файлов в MinIO.
// Формируется как <эндпоинт>/<бакет>/
const minioBaseURL = "http://localhost:9000/glotto/"

// Локальные дефолтные файлы (используются, если ключ пустой)
const defaultImage = "/static/img/default.jpg"
const defaultVideo = "/static/video/default.mp4"

// imageURLFor — возвращает полный URL картинки из MinIO
// или локальный дефолт, если ключ пустой
func imageURLFor(key string) string {
    if key == "" {
        return defaultImage
    }
    if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
        return key
    }
    if strings.HasPrefix(key, "/static/") {
        return key
    }
    return minioBaseURL + key
}

// videoURLFor — возвращает полный URL видео из MinIO
// или локальный дефолт, если ключ пустой
func videoURLFor(key string) string {
    if key == "" {
        return defaultVideo
    }
    if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
        return key
    }
    if strings.HasPrefix(key, "/static/") {
        return key
    }
    return minioBaseURL + key
}

func formatYear(t time.Time) string {
	y := t.Year()
	if y < 0 {
		return strconv.Itoa(-y) + " г. до н.э."
	}
	return strconv.Itoa(y) + " г."
}

func formatYearInt(y int) string {
	if y < 0 {
		return strconv.Itoa(-y) + " г. до н.э."
	}
	return strconv.Itoa(y) + " г."
}

// --- GET /languages ---
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

	c.HTML(http.StatusOK, "swadesh_feed.html", gin.H{
		"service":  lang,
		"likes":    h.Repository.GetLikeCount(lang.ID),
		"imageURL": imageURLFor(lang.ImageKey),
		"videoURL": videoURLFor(lang.VideoKey),
		"dateStr":  formatYear(lang.DivergenceDate),
	})
}

// --- GET /languages/new ---
func (h *Handler) AddHandler(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if draft != nil && draft.ID != 0 {
		c.Redirect(302, "/swadesh-lists/publish")
		return
	}

	c.HTML(http.StatusOK, "swadesh_add_step1.html", gin.H{
		"defaultImage": defaultImage,
		"defaultVideo": defaultVideo,
	})
}

// --- GET /languages/catalog ---
func (h *Handler) GridHandler(c *gin.Context) {
    dateStr := c.Query("date")

    var langs []ds.Language
    var err error

    if dateStr == "" {
        langs, err = h.Repository.GetAllPublished()
    } else {
        // dateStr = "1500-01-01" → парсим, берём .Year()
        t, e := time.Parse("2006-01-02", dateStr)
        if e == nil {
            langs, err = h.Repository.GetPublishedFilteredByYear(t.Year())
        } else {
            langs, err = h.Repository.GetAllPublished()
        }
    }
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
            ImageURL: imageURLFor(l.ImageKey),
            DateStr:  formatYear(l.DivergenceDate),
            Likes:    h.Repository.GetLikeCount(l.ID),
        })
    }

    c.HTML(http.StatusOK, "swadesh_grid.html", gin.H{
        "cards":  cards,
        "filter": dateStr,
    })
}

// --- POST /languages/new ---
// Создание черновика: сохраняем название, ссылку на картинку и видео.
func (h *Handler) CreateDraftHandler(c *gin.Context) {
	title := c.PostForm("title")
	imageURL := c.PostForm("image_url")
	videoURL := c.PostForm("video_url")

	if imageURL == "" {
		imageURL = defaultImage
	}
	if videoURL == "" {
		videoURL = defaultVideo
	}

	lang := &ds.Language{
		Title:          title,
		Description:    "",
		Status:         "draft",
		ImageKey:       imageURL,
		VideoKey:       videoURL,
		MatchPercent:   0,
		DivergenceDate: time.Now(),
		CreatorID:      currentUserID,
	}

	if err := h.Repository.CreateLanguage(lang); err != nil {
		h.errorHandler(c, http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/swadesh-lists/publish")
}

// --- GET /languages/publish ---
func (h *Handler) PublishFormHandler(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(currentUserID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if draft == nil {
		c.Redirect(302, "/swadesh-lists/new")
		return
	}

	c.HTML(http.StatusOK, "swadesh_add_step2.html", gin.H{
		"service":  draft,
		"imageURL": imageURLFor(draft.ImageKey),
		"videoURL": videoURLFor(draft.VideoKey),
	})
}

// --- POST /languages/publish ---
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

	c.Redirect(http.StatusFound, "/swadesh-lists/catalog")
}

// --- POST /languages/delete ---
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

	c.Redirect(http.StatusFound, "/swadesh-lists/catalog")
}
