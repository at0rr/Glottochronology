package ds

import (
	"strconv"
	"strings"
)

// Локальные дефолтные файлы (используются, если ключ пустой)
const defaultImage = "/static/img/default.jpg"
const defaultVideo = "/static/video/default.mp4"

// formatYearStr форматирует год в человекочитаемую строку
func formatYearStr(t interface{ Year() int }) string {
	y := t.Year()
	if y < 0 {
		return strconv.Itoa(-y) + " г. до н.э."
	}
	return strconv.Itoa(y) + " г."
}

// buildImageURL — умное формирование URL картинки.
//   - пусто → локальный дефолт
//   - http:// или https:// → вернуть как есть
//   - /static/ → вернуть как есть (локальный файл)
//   - иначе → publicURL + "/" + key (ключ MinIO)
func buildImageURL(key, publicURL string) string {
	if key == "" {
		return defaultImage
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	if strings.HasPrefix(key, "/static/") {
		return key
	}
	return publicURL + "/" + key
}

// buildVideoURL — то же самое для видео
func buildVideoURL(key, publicURL string) string {
	if key == "" {
		return defaultVideo
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	if strings.HasPrefix(key, "/static/") {
		return key
	}
	return publicURL + "/" + key
}

// LanguageSerializer — сериализатор для списка и ленты
type LanguageSerializer struct {
	ID             uint   `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Status         string `json:"status"`
	ImageURL       string `json:"image_url"`
	VideoURL       string `json:"video_url"`
	MatchPercent   int    `json:"match_percent"`
	DivergenceDate string `json:"divergence_date"`
	DateCreate     string `json:"date_create"`
	DateFormation  string `json:"date_formation"`
	CreatorID      uint   `json:"creator_id"`
	Likes          int64  `json:"likes"`
}

func NewLanguageSerializer(l Language, likes int64, publicURL string) LanguageSerializer {
	return LanguageSerializer{
		ID:             l.ID,
		Title:          l.Title,
		Description:    l.Description,
		Status:         l.Status,
		ImageURL:       buildImageURL(l.ImageKey, publicURL),
		VideoURL:       buildVideoURL(l.VideoKey, publicURL),
		MatchPercent:   l.MatchPercent,
		DivergenceDate: formatYearStr(l.DivergenceDate),
		DateCreate:     l.DateCreate.Format("2006-01-02 15:04:05"),
		DateFormation:  l.DateFormation.Format("2006-01-02 15:04:05"),
		CreatorID:      l.CreatorID,
		Likes:          likes,
	}
}

// FeedSerializer — сериализатор для ленты (добавляет поле next_id)
type FeedSerializer struct {
	LanguageSerializer
	NextID *uint `json:"next_id,omitempty"`
}

func NewFeedSerializer(l Language, likes int64, publicURL string, nextID *uint) FeedSerializer {
	return FeedSerializer{
		LanguageSerializer: NewLanguageSerializer(l, likes, publicURL),
		NextID:             nextID,
	}
}

// UserSerializer — сериализатор пользователя
type UserSerializer struct {
	ID        uint   `json:"id"`
	Login     string `json:"login"`
}

func NewUserSerializer(u User) UserSerializer {
	return UserSerializer{
		ID:        u.ID,
		Login:     u.Login,
	}
}

// ErrorResponse — общий формат ошибки
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewError(message string) ErrorResponse {
	return ErrorResponse{Status: "error", Message: message}
}

// Success — общий формат успеха
func Success(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"status": "success",
		"data":   data,
	}
}

// SuccessResponse — общий формат успеха
type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// LikeResponse — ответ на лайк
type LikeResponse struct {
	Likes int64 `json:"likes"`
	Liked bool  `json:"liked"`
}
