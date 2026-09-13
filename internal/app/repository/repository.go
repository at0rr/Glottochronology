package repository

import (
	"fmt"
	"strconv"
	"time"
)

// Service — структура услуги (списка Сводеша для языка)
type Service struct {
	ID            int
	Title         string    // Название языка
	Description   string    // Описание
	Date          time.Time // Дата фиксации/расхождения
	MatchPercent  int       // Характерное число (процент совпадений) — для фильтрации
	Status        string    // "draft", "published", "deleted"
	ImageKey      string    // Ключ изображения в MinIO
	VideoKey      string    // Ключ видео в MinIO
	Likes         []string  // Массив ID пользователей, лайкнувших услугу
}

// Repository — хранилище данных
type Repository struct {
	services []Service
}

// NewRepository создаёт репозиторий и заполняет тестовыми данными
func NewRepository() (*Repository, error) {
	services := []Service{
		{
			ID:           1,
			Title:        "Русский",
			Description:  "Используется для сравнения с другими славянскими языками.",
			Date:         time.Date(1500, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 85,
			Status:       "published",
			ImageKey:     "russian.jpg",
			VideoKey:     "russian.mp4",
			Likes:        []string{"user1", "user3", "user5", "user7", "user9", "user11", "user13", "user15", "user17", "user19", "user21", "user23"},
		},
		{
			ID:           2,
			Title:        "Английский",
			Description:  "Позволяет оценить степень родства с другими германскими языками.",
			Date:         time.Date(450, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 72,
			Status:       "published",
			ImageKey:     "english.jpg",
			VideoKey:     "english.mp4",
			Likes:        []string{"user1", "user2", "user4", "user6", "user8", "user10", "user12", "user14"},
		},
		{
			ID:           3,
			Title:        "Немецкий",
			Description:  "Используется в сравнительно-историческом языкознании.",
			Date:         time.Date(500, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 68,
			Status:       "published",
			ImageKey:     "german.jpg",
			VideoKey:     "germany.mp4",
			Likes:        []string{"user1", "user3", "user5", "user7", "user9"},
		},
		{
			ID:           4,
			Title:        "Французский",
			Description:  "Показывает эволюцию лексики от латинских корней.",
			Date:         time.Date(600, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 65,
			Status:       "published",
			ImageKey:     "french.jpg",
			VideoKey:     "france.mp4",
			Likes:        []string{"user2", "user4", "user6", "user8", "user10", "user12", "user14"},
		},
		{
			ID:           5,
			Title:        "Китайский",
			Description:  "Используется для изучения древних контактов с соседними языками.",
			Date:         time.Date(-2000, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 40,
			Status:       "published",
			ImageKey:     "chinese.jpg",
			VideoKey:     "china.mp4",
			Likes:        []string{"user1", "user3", "user5"},
		},
		{
			ID:           6,
			Title:        "Японский",
			Description:  "Список Сводеша для японского языка.",
			Date:         time.Date(300, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 35,
			Status:       "published",
			ImageKey:     "japanese.jpg",
			VideoKey:     "japan.mp4",
			Likes:        []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7"},
		},
		{
			ID:           7,
			Title:        "Санскрит",
			Description:  "Черновик услуги для санскрита. Требуется уточнить список Сводеша и заполнить поля по глоттохронологии. Используется для сравнения с другими индоевропейскими языками.",
			Date:         time.Date(-1500, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 0,
			Status:       "draft",
			ImageKey:     "russian.jpg",
			VideoKey:     "russian.mp4",
			Likes:        []string{},
		},
		{
			ID:           8,
			Title:        "Латынь",
			Description:  "Удалённая услуга. В интерфейсе не отображается.",
			Date:         time.Date(-100, 1, 1, 0, 0, 0, 0, time.UTC),
			MatchPercent: 0,
			Status:       "deleted",
			ImageKey:     "",
			VideoKey:     "",
			Likes:        []string{},
		},
	}
	return &Repository{services: services}, nil
}

// GetAllPublished возвращает все опубликованные услуги
func (r *Repository) GetAllPublished() []Service {
	var result []Service
	for _, s := range r.services {
		if s.Status == "published" {
			result = append(result, s)
		}
	}
	return result
}

// GetByID возвращает услугу по ID (кроме удалённых)
func (r *Repository) GetByID(id int) (Service, error) {
	for _, s := range r.services {
		if s.ID == id && s.Status != "deleted" {
			return s, nil
		}
	}
	return Service{}, fmt.Errorf("услуга не найдена или удалена")
}

// GetNextAfterID возвращает следующую опубликованную услугу после указанного ID
func (r *Repository) GetNextAfterID(id int) (Service, error) {
	published := r.GetAllPublished()
	found := false
	for _, s := range published {
		if found {
			return s, nil
		}
		if s.ID == id {
			found = true
		}
	}
	return Service{}, fmt.Errorf("следующая услуга не найдена")
}

// GetDraft возвращает черновик
func (r *Repository) GetDraft() (Service, error) {
	for _, s := range r.services {
		if s.Status == "draft" {
			return s, nil
		}
	}
	return Service{}, fmt.Errorf("черновик не найден")
}

// GetPublishedFilteredByDate фильтрует опубликованные услуги по году расхождения
// dateFilter — строка, содержащая год (например, "1500" или "-2000")
func (r *Repository) GetPublishedFilteredByDate(dateFilter string) []Service {
	published := r.GetAllPublished()
	if dateFilter == "" {
		return published
	}
	filterVal, err := strconv.Atoi(dateFilter)
	if err != nil {
		return published
	}
	var result []Service
	for _, s := range published {
		if s.Date.Year() == filterVal {
			result = append(result, s)
		}
	}
	return result
}
