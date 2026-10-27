package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"glottochronology/internal/app/ds"
	"glottochronology/internal/app/middleware"
)

// ============ ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ============

func formatYear(t time.Time) string {
	y := t.Year()
	if y < 0 {
		return strconv.Itoa(-y) + " г. до н.э."
	}
	return strconv.Itoa(y) + " г."
}

func imageURLFor(key string) string {
	if key == "" {
		return "/static/img/default.jpg"
	}
	return "http://localhost:9000/glotto/" + key
}

func videoURLFor(key string) string {
	if key == "" {
		return "/static/video/default.mp4"
	}
	return "http://localhost:9000/glotto/" + key
}

// ============ API-МЕТОДЫ (JSON для Postman/Swagger) ============

// APIGetLanguages godoc
// @Summary      Список опубликованных языков
// @Description  Возвращает список опубликованных языков с опциональным фильтром
// @Tags         languages
// @Produce      json
// @Param        mode  query string false "Режим фильтра: year | decade | century"
// @Param        value query int    false "Значение фильтра (год)"
// @Success      200   {array}   ds.LanguageSerializer
// @Failure      500   {object}  ds.ErrorResponse
// @Router       /api/swadesh-lists [get]
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
	userID := middleware.CurrentUserID(c)

	result := make([]ds.LanguageSerializer, 0, len(langs))
	for _, l := range langs {
		likes := h.Repository.GetLikeCount(l.ID)
		isLiked := h.Repository.HasLike(userID, l.ID)
		result = append(result, ds.NewLanguageSerializer(l, likes, userID, isLiked, pubURL))
	}

	c.JSON(http.StatusOK, result)
}

// APIGetFeed godoc
// @Summary      Лента языков
// @Description  Возвращает первую опубликованную услугу или конкретную по id (в query). С ?next=true возвращает следующую.
// @Tags         languages
// @Produce      json
// @Param        id   query int  false "ID языка"
// @Param        next query bool false "Если true — следующая за указанной"
// @Success      200  {object}  ds.FeedSerializer
// @Failure      400  {object}  ds.ErrorResponse
// @Failure      404  {object}  ds.ErrorResponse
// @Router       /api/swadesh-lists/feed [get]
func (h *Handler) APIGetFeed(c *gin.Context) {
	idStr := c.Query("id")
	pubURL := h.Repository.PublicURL()
	userID := middleware.CurrentUserID(c)

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
		isLiked := h.Repository.HasLike(userID, lang.ID)
		c.JSON(http.StatusOK, ds.NewFeedSerializer(lang, h.Repository.GetLikeCount(lang.ID), userID, isLiked, pubURL, nextID))
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

	isLiked := h.Repository.HasLike(userID, lang.ID)
	c.JSON(http.StatusOK, ds.NewFeedSerializer(*lang, h.Repository.GetLikeCount(lang.ID), userID, isLiked, pubURL, nextID))
}

// APIGetDraft godoc
// @Summary      Черновик текущего пользователя
// @Tags         languages
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ds.LanguageSerializer
// @Router       /api/swadesh-lists/draft [get]
func (h *Handler) APIGetDraft(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	draft, err := h.Repository.GetDraftByCreator(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}
	if draft == nil {
		c.JSON(http.StatusNotFound, ds.NewError("черновик не найден"))
		return
	}

	pubURL := h.Repository.PublicURL()
	isLiked := h.Repository.HasLike(userID, draft.ID)
	c.JSON(http.StatusOK, ds.NewLanguageSerializer(*draft, h.Repository.GetLikeCount(draft.ID), userID, isLiked, pubURL))
}

// APICreateLanguage godoc
// @Summary      Создать черновик языка
// @Tags         languages
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        title formData string true  "Название языка"
// @Param        image formData file   false "Изображение"
// @Param        video formData file   false "Видео"
// @Success      201  {object}  ds.LanguageSerializer
// @Router       /api/swadesh-lists [post]
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

	userID := middleware.CurrentUserID(c)
	existing, _ := h.Repository.GetDraftByCreator(userID)
	if existing != nil {
		c.JSON(http.StatusConflict, ds.NewError("у пользователя уже есть черновик"))
		return
	}

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
		Status:         "draft",
		ImageKey:       imageKey,
		VideoKey:       videoKey,
		DivergenceDate: time.Now(),
		CreatorID:      userID,
	}

	if err := h.Repository.CreateDraft(lang); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ds.NewLanguageSerializer(*lang, 0, userID, false, h.Repository.PublicURL()))
}

// APIPublishLanguage godoc
// @Summary      Опубликовать услугу
// @Tags         languages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "ID услуги"
// @Router       /api/swadesh-lists/{id}/publish [put]
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

	userID := middleware.CurrentUserID(c)
	isLiked := h.Repository.HasLike(userID, lang.ID)
	c.JSON(http.StatusOK, ds.NewLanguageSerializer(*lang, h.Repository.GetLikeCount(lang.ID), userID, isLiked, h.Repository.PublicURL()))
}

// APIDeleteLanguage godoc
// @Summary      Логически удалить услугу
// @Tags         languages
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID услуги"
// @Router       /api/swadesh-lists/{id} [delete]
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

	c.JSON(http.StatusOK, gin.H{})
}

// APILikeLanguage godoc
// @Summary      Поставить/снять лайк
// @Tags         languages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "ID услуги"
// @Router       /api/swadesh-lists/{id}/like [post]
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
	if err := c.ShouldBindJSON(&body); err != nil || (body.Like != 0 && body.Like != 1) {
		c.JSON(http.StatusBadRequest, ds.NewError("поле like должно быть 0 или 1"))
		return
	}

	userID := middleware.CurrentUserID(c)
	if err := h.Repository.SetLike(userID, uint(id), body.Like); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ds.Success(map[string]interface{}{
		"likes": h.Repository.GetLikeCount(uint(id)),
		"liked": h.Repository.HasLike(userID, uint(id)),
	}))
}

// ============ HTML-МЕТОДЫ (Для браузера) ============

// FeedHandler — страница ленты
func (h *Handler) FeedHandler(c *gin.Context) {
	idStr := c.Query("id")
	var lang *ds.Language
	var err error

	if idStr == "" {
		all, _ := h.Repository.GetAllPublished("", 0, false)
		if len(all) > 0 {
			lang = &all[0]
		}
	} else {
		id, _ := strconv.Atoi(idStr)
		if c.Query("next") == "true" {
			next, _ := h.Repository.GetNextAfterID(id)
			if next != nil {
				lang = next
			} else {
				lang, _ = h.Repository.GetByID(id)
			}
		} else {
			lang, err = h.Repository.GetByID(id)
		}
	}

	if err != nil || lang == nil {
		c.String(http.StatusNotFound, "Услуга не найдена")
		return
	}

	c.HTML(http.StatusOK, "feed.html", gin.H{
		"service":  lang,
		"likes":    h.Repository.GetLikeCount(lang.ID),
		"imageURL": imageURLFor(lang.ImageKey),
		"videoURL": videoURLFor(lang.VideoKey),
		"dateStr":  formatYear(lang.DivergenceDate),
	})
}

// GridHandler — страница плитки
// GridHandler — страница плитки
func (h *Handler) GridHandler(c *gin.Context) {
	dateStr := c.Query("date")
	mode := "year"
	value := 0
	hasFilter := false

	// Если дата выбрана, парсим её и берем только год
	if dateStr != "" {
		t, e := time.Parse("2006-01-02", dateStr)
		if e == nil {
			value = t.Year()
			hasFilter = true
		}
	}

	langs, err := h.Repository.GetAllPublished(mode, value, hasFilter)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
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

	c.HTML(http.StatusOK, "grid.html", gin.H{
		"cards":  cards,
		"filter": dateStr,
	})
}

// ============ HTML-МЕТОДЫ ДЛЯ ДОБАВЛЕНИЯ И ПУБЛИКАЦИИ ============

// AddHandler — страница добавления (шаг 1)
func (h *Handler) AddHandler(c *gin.Context) {
	// Для HTML-страниц используем захардкоженного пользователя (id=1),
	// так как на этих страницах нет JWT-мидлвари.
	draft, err := h.Repository.GetDraftByCreator(1)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if draft != nil && draft.ID != 0 {
		c.Redirect(302, "/swadesh-lists/publish")
		return
	}

	c.HTML(http.StatusOK, "add-step1.html", gin.H{})
}

// PublishFormHandler — страница публикации (шаг 2)
func (h *Handler) PublishFormHandler(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(1)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if draft == nil {
		c.Redirect(302, "/swadesh-lists/new")
		return
	}

	c.HTML(http.StatusOK, "add-step2.html", gin.H{
		"service":  draft,
		"imageURL": imageURLFor(draft.ImageKey),
		"videoURL": videoURLFor(draft.VideoKey),
	})
}

// CreateDraftHandler — POST обработка создания черновика из HTML формы
func (h *Handler) CreateDraftHandler(c *gin.Context) {
	title := c.PostForm("title")

	lang := &ds.Language{
		Title:          title,
		Status:         "draft",
		DivergenceDate: time.Now(),
		CreatorID:      1,
	}

	if err := h.Repository.CreateDraft(lang); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/swadesh-lists/publish")
}

// PublishHandler — POST обработка публикации из HTML формы
func (h *Handler) PublishHandler(c *gin.Context) {
	idStr := c.PostForm("id")
	id, _ := strconv.Atoi(idStr)

	title := c.PostForm("title")
	description := c.PostForm("description")
	matchStr := c.PostForm("match_percent")
	divDate := c.PostForm("divergence_date")
	match, _ := strconv.Atoi(matchStr)

	if err := h.Repository.Publish(uint(id), title, description, match, divDate); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/swadesh-lists/catalog")
}
