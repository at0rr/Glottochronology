package repository

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db          *gorm.DB
	minio       *minio.Client
	minioBucket string
	minioPubURL string
}

type Settings struct {
	PostgresDSN    string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	MinioPublicURL string
}

func New(s Settings) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(s.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := minio.New(s.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.MinioAccessKey, s.MinioSecretKey, ""),
		Secure: s.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		minio:       minioClient,
		minioBucket: s.MinioBucket,
		minioPubURL: s.MinioPublicURL + "/" + s.MinioBucket,
	}, nil
}

// PublicURL — базовый URL для файлов
func (r *Repository) PublicURL() string {
	return r.minioPubURL
}
