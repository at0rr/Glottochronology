package repository

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// UploadFile — загружает файл в MinIO, возвращает ключ
func (r *Repository) UploadFile(header *multipart.FileHeader, prefix string) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	// Определяем Content-Type
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", fmt.Errorf("не удалось прочитать файл: %w", err)
	}
	contentType := http.DetectContentType(buffer)
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("ошибка возврата позиции: %w", err)
	}

	// Генерируем ключ на латинице
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}
	key := fmt.Sprintf("%s_%s%s", prefix, uuid.New().String(), ext)

	_, err = r.minio.PutObject(
		context.Background(),
		r.minioBucket,
		key,
		file,
		header.Size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки в MinIO: %w", err)
	}

	return key, nil
}

// isImage — определяет, картинка ли это
func isImage(contentType string) bool {
	images := []string{
		"image/jpeg", "image/jpg", "image/png",
		"image/gif", "image/webp", "image/bmp",
	}
	for _, t := range images {
		if contentType == t {
			return true
		}
	}
	return false
}

// isVideo — определяет, видео ли это
func isVideo(contentType string) bool {
	videos := []string{
		"video/mp4", "video/webm", "video/quicktime",
		"video/x-msvideo", "video/mpeg",
	}
	for _, t := range videos {
		if contentType == t {
			return true
		}
	}
	return false
}
