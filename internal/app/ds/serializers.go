package ds

import (
	"strconv"
	"strings"
)

const defaultImage = "/static/img/default.jpg"
const defaultVideo = "/static/video/default.mp4"

func formatYearStr(t interface{ Year() int }) string {
	y := t.Year()
	if y < 0 {
		return strconv.Itoa(-y) + " г. до н.э."
	}
	return strconv.Itoa(y) + " г."
}

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
	IsCreator      bool   `json:"is_creator"`
	IsLiked        bool   `json:"is_liked"`
	Attributes []AttributeValue `json:"attributes"`
}

// ОБНОВЛЕНО: добавлены currentUserID и isLiked
func NewLanguageSerializer(l Language, likes int64, currentUserID uint, isLiked bool, publicURL string) LanguageSerializer {
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
		IsCreator:      l.CreatorID == currentUserID,
		IsLiked:        isLiked,
		Attributes:     l.Attributes, // ВОТ ТУТ передаем атрибуты из модели в ответ
	}
}

type FeedSerializer struct {
	LanguageSerializer
	NextID *uint `json:"next_id,omitempty"`
}

// ОБНОВЛЕНО: добавлены currentUserID и isLiked
func NewFeedSerializer(l Language, likes int64, currentUserID uint, isLiked bool, publicURL string, nextID *uint) FeedSerializer {
	return FeedSerializer{
		LanguageSerializer: NewLanguageSerializer(l, likes, currentUserID, isLiked, publicURL),
		NextID:             nextID,
	}
}

type UserSerializer struct {
	ID        uint   `json:"id"`
	Login     string `json:"login"`
}

func NewUserSerializer(u User) UserSerializer {
	return UserSerializer{
		ID:    u.ID,
		Login: u.Login,
	}
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewError(message string) ErrorResponse {
	return ErrorResponse{Status: "error", Message: message}
}

func Success(data interface{}) interface{} {
	return data
}

type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type LikeResponse struct {
	Likes int64 `json:"likes"`
	Liked bool  `json:"liked"`
}
