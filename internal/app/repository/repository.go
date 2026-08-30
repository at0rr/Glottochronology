package repository

import "fmt"
import "strings"

// Language — структура языка (списка Сводеша)
type Language struct {
	ID    int
	Title string // название языка
}

// Repository — хранилище данных (пока в памяти)
type Repository struct {
	// можно добавить поле для хранения данных, но пока обойдёмся
}

// NewRepository создаёт новый репозиторий
func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// GetLanguages возвращает все языки (имитация БД)
func (r *Repository) GetLanguages() ([]Language, error) {
	languages := []Language{
		{ID: 1, Title: "Русский"},
		{ID: 2, Title: "Английский"},
		{ID: 3, Title: "Немецкий"},
	}
	if len(languages) == 0 {
		return nil, fmt.Errorf("список языков пуст")
	}
	return languages, nil
}

// GetLanguageByID возвращает язык по ID
func (r *Repository) GetLanguageByID(id int) (Language, error) {
	languages, err := r.GetLanguages()
	if err != nil {
		return Language{}, err
	}
	for _, lang := range languages {
		if lang.ID == id {
			return lang, nil
		}
	}
	return Language{}, fmt.Errorf("язык с ID %d не найден", id)
}

// SearchLanguagesByTitle ищет языки, содержащие подстроку (без учёта регистра)
func (r *Repository) SearchLanguagesByTitle(query string) ([]Language, error) {
	languages, err := r.GetLanguages()
	if err != nil {
		return []Language{}, err
	}
	var result []Language
	for _, lang := range languages {
		if strings.Contains(strings.ToLower(lang.Title), strings.ToLower(query)) {
			result = append(result, lang)
		}
	}
	return result, nil
}
