package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"glottochronology/internal/app/ds"
	"glottochronology/internal/app/middleware"
)

// ============ API-обработчики ============

// APIGetLanguages godoc
// @Summary      Список опубликованных языков
// @Description  Возвращает список опубликованных языков с опциональным фильтром по году/десятилетию/столетию
// @Tags         languages
// @Produce      json
// @Param        mode  query string false "Режим фильтра: year | decade | century"
// @Param        value query int    false "Значение фильтра (год)"
// @Success      200   {array}   ds.LanguageSerializer
// @Failure      500   {object}  ds.ErrorResponse
// @Router       /api/languages [get]
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

// APIGetFeed godoc
// @Summary      Лента языков
// @Description  Возвращает первую опубликованную услугу или конкретную по id. С ?next=true возвращает следующую за указанной
// @Tags         languages
// @Produce      json
// @Param        id   path  int  false "ID языка"
// @Param        next query bool false "Если true — следующая за указанной"
// @Success      200  {object}  ds.FeedSerializer
// @Failure      400  {object}  ds.ErrorResponse
// @Failure      404  {object}  ds.ErrorResponse
// @Router       /api/languages/{id} [get]
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

// APIGetDraft godoc
// @Summary      Черновик текущего пользователя
// @Description  Возвращает черновик авторизованного пользователя (не более одного)
// @Tags         languages
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  ds.LanguageSerializer
// @Failure      401  {object}  ds.ErrorResponse
// @Failure      404  {object}  ds.ErrorResponse
// @Router       /api/languages/draft [get]
func (h *Handler) APIGetDraft(c *gin.Context) {
	draft, err := h.Repository.GetDraftByCreator(middleware.CurrentUserID(c))
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

// APICreateLanguage godoc
// @Summary      Создать черновик языка
// @Description  Создаёт новый черновик услуги с загрузкой изображения и видео в MinIO
// @Tags         languages
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        title formData string true  "Название языка"
// @Param        image formData file   false "Изображение"
// @Param        video formData file   false "Видео"
// @Success      201  {object}  ds.LanguageSerializer
// @Failure      400  {object}  ds.ErrorResponse
// @Failure      401  {object}  ds.ErrorResponse
// @Failure      409  {object}  ds.ErrorResponse
// @Router       /api/languages [post]
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

	existing, _ := h.Repository.GetDraftByCreator(middleware.CurrentUserID(c))
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
		Description:    "",
		Status:         "draft",
		ImageKey:       imageKey,
		VideoKey:       videoKey,
		MatchPercent:   0,
		DivergenceDate: time.Now(),
		CreatorID:      middleware.CurrentUserID(c),
	}

	if err := h.Repository.CreateDraft(lang); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ds.Success(
		ds.NewLanguageSerializer(*lang, 0, h.Repository.PublicURL()),
	))
}

// APIPublishLanguage godoc
// @Summary      Опубликовать услугу
// @Description  Меняет статус черновика на published и заполняет поля
// @Tags         languages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "ID услуги"
// @Success      200  {object}  ds.LanguageSerializer
// @Failure      400  {object}  ds.ErrorResponse
// @Failure      401  {object}  ds.ErrorResponse
// @Router       /api/languages/{id}/publish [put]
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

	c.JSON(http.StatusOK, ds.Success(
		ds.NewLanguageSerializer(*lang, h.Repository.GetLikeCount(lang.ID), h.Repository.PublicURL()),
	))
}

// APIDeleteLanguage godoc
// @Summary      Логически удалить услугу
// @Description  Меняет статус услуги на deleted через SQL UPDATE (soft delete)
// @Tags         languages
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "ID услуги"
// @Success      200  {object}  ds.SuccessResponse
// @Failure      401  {object}  ds.ErrorResponse
// @Failure      404  {object}  ds.ErrorResponse
// @Router       /api/languages/{id} [delete]
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

// APILikeLanguage godoc
// @Summary      Поставить/снять лайк
// @Description  Поле like = 1 ставит лайк, like = 0 снимает
// @Tags         languages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path int true "ID услуги"
// @Success      200  {object}  ds.LikeResponse
// @Failure      400  {object}  ds.ErrorResponse
// @Failure      401  {object}  ds.ErrorResponse
// @Failure      404  {object}  ds.ErrorResponse
// @Router       /api/languages/{id}/like [post]
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

	lang, _ := h.Repository.GetByID(id)
	if lang == nil {
		c.JSON(http.StatusNotFound, ds.NewError("услуга не найдена"))
		return
	}

	if err := h.Repository.SetLike(middleware.CurrentUserID(c), uint(id), body.Like); err != nil {
		c.JSON(http.StatusInternalServerError, ds.NewError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ds.Success(map[string]interface{}{
		"likes": h.Repository.GetLikeCount(uint(id)),
		"liked": h.Repository.HasLike(middleware.CurrentUserID(c), uint(id)),
	}))
}
