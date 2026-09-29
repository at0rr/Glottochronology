package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"glottochronology/internal/app/ds"
	"glottochronology/internal/app/singleton"
)

// ============ API-обработчики ============

// GET /api/languages?mode=year&value=500
func (h *Handler) APIGetLanguages(c *gin.Context) {
	mode := c.DefaultQuery("mode", "year")
	valueStr := c.Query("value")
	hasFilter := valueStr != ""
	value := 0
	if hasFilter {
		value, _ = strconv.Atoi(valueStr)
	}

	langs, err := h.Repository.GetAllPublished(mode, value, hasFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	pubURL := h.Repository.PublicURL()
	result := make([]ds.LanguageSerializer, 0, len(langs))
	for _, l := range langs {
		likes := h.Repository.GetLikeCount(l.ID)
		result = append(result, ds.NewLanguageSerializer(l, likes, pubURL))
	}

	c.JSON(http.StatusOK, ds.Success(result))
}

// GET /api/languages/:id
// GET /api/languages/:id?next=true — следующая за указанной
// GET /api/languages/feed — без id, первая опубликованная
func (h *Handler) APIGetFeed(c *gin.Context) {
	idStr := c.Param("id")
	pubURL := h.Repository.PublicURL()

	// без id → первая опубликованная
	if idStr == "" {
		all, err := h.Repository.GetAllPublished("", 0, false)
		if err != nil || len(all) == 0 {
			c.JSON(http.StatusNotFound, ds.NewError("нет опубликованных услуг"))
			return
		}
		lang := all[0]
		var nextID *uint
		if len(all) > 1 {
			nextID = &all[1].ID
		}
		c.JSON(http.StatusOK, ds.Success(
			ds.NewFeedSerializer(lang, h.Repository.GetLikeCount(lang.ID), pubURL, nextID),
		))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверный id"))
		return
	}

	var lang *ds.Language
	if c.Query("next") == "true" {
		next, _ := h.Repository.GetNextAfterID(id)
		if next == nil {
			lang, _ = h.Repository.GetByID(id)
		} else {
			lang = next
		}
	} else {
		lang, _ = h.Repository.GetByID(id)
	}

	if lang == nil {
		c.JSON(http.StatusNotFound, ds.NewError("услуга не найдена"))
		return
	}

	var nextID *uint
	if next, _ := h.Repository.GetNextAfterID(int(lang.ID)); next != nil {
		nextID = &next.ID
	}

	c.JSON(http.StatusOK, ds.Success(
		ds.NewFeedSerializer(*lang, h.Repository.GetLikeCount(lang.ID), pubURL, nextID),
	))
}

// GET /api/languages/draft — черновик текущего пользователя
func (h *Handler) APIGetDraft(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(singleton.CurrentUserID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}
	if draft == nil {
		c.JSON(http.StatusNotFound, ds.NewError("черновик не найден"))
		return
	}

	pubURL := h.Repository.PublicURL()
	c.JSON(http.StatusOK, ds.Success(
		ds.NewLanguageSerializer(*draft, h.Repository.GetLikeCount(draft.ID), pubURL),
	))
}

// POST /api/languages — создание черновика (multipart/form-data)
// Поля: title, image (file), video (file)
func (h *Handler) APICreateLanguage(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("ошибка разбора формы: "+err.Error()))
		return
	}

	title := c.PostForm("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, ds.NewError("поле title обязательно"))
		return
	}

	// Запрещаем создавать второй черновик
	existing, _ := h.Repository.GetDraftByCreator(singleton.CurrentUserID())
	if existing != nil {
		c.JSON(http.StatusConflict, ds.NewError("у пользователя уже есть черновик"))
		return
	}

	// Загружаем картинку
	var imageKey, videoKey string
	if header, err := c.FormFile("image"); err == nil {
		key, err := h.Repository.UploadFile(header, "image")
		if err != nil {
			c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
			return
		}
		imageKey = key
	}
	if header, err := c.FormFile("video"); err == nil {
		key, err := h.Repository.UploadFile(header, "video")
		if err != nil {
			c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
			return
		}
		videoKey = key
	}

	lang := &ds.Language{
		Title:          title,
		Description:    "",
		Status:         "draft",
		ImageKey:       imageKey,
		VideoKey:       videoKey,
		MatchPercent:   0,
		DivergenceDate: time.Now(),
		CreatorID:      singleton.CurrentUserID(),
	}

	if err := h.Repository.CreateDraft(lang); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ds.Success(
		ds.NewLanguageSerializer(*lang, 0, h.Repository.PublicURL()),
	))
}

// PUT /api/languages/:id/publish — публикация (JSON или form)
func (h *Handler) APIPublishLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверный id"))
		return
	}

	var body struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		MatchPercent   int    `json:"match_percent"`
		DivergenceDate string `json:"divergence_date"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		// fallback на form-data
		body.Title = c.PostForm("title")
		body.Description = c.PostForm("description")
		body.MatchPercent, _ = strconv.Atoi(c.PostForm("match_percent"))
		body.DivergenceDate = c.PostForm("divergence_date")
	}

	if err := h.Repository.Publish(uint(id), body.Title, body.Description, body.MatchPercent, body.DivergenceDate); err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError(err.Error()))
		return
	}

	lang, _ := h.Repository.GetByID(id)
	if lang == nil {
		c.JSON(http.StatusInternalServerError, ds.NewError("не удалось получить услугу"))
		return
	}

	c.JSON(http.StatusOK, ds.Success(
		ds.NewLanguageSerializer(*lang, h.Repository.GetLikeCount(lang.ID), h.Repository.PublicURL()),
	))
}

// DELETE /api/languages/:id — soft delete (SQL UPDATE)
func (h *Handler) APIDeleteLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверный id"))
		return
	}

	if err := h.Repository.SoftDelete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ds.Success(map[string]string{"message": "услуга удалена"}))
}

// POST /api/languages/:id/like — { "like": 0 | 1 }
func (h *Handler) APILikeLanguage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("неверный id"))
		return
	}

	var body struct {
		Like int `json:"like"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ds.NewError("тело должно содержать поле like"))
		return
	}
	if body.Like != 0 && body.Like != 1 {
		c.JSON(http.StatusBadRequest, ds.NewError("поле like должно быть 0 или 1"))
		return
	}

	// Проверим, что услуга есть и не удалена
	lang, _ := h.Repository.GetByID(id)
	if lang == nil {
		c.JSON(http.StatusNotFound, ds.NewError("услуга не найдена"))
		return
	}

	if err := h.Repository.SetLike(singleton.CurrentUserID(), uint(id), body.Like); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ds.Success(map[string]interface{}{
		"likes": h.Repository.GetLikeCount(uint(id)),
		"liked": h.Repository.HasLike(singleton.CurrentUserID(), uint(id)),
	}))
}
