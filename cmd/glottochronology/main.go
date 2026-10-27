package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "glottochronology/docs"

	"glottochronology/internal/app/dsn"
	"glottochronology/internal/app/handler"
	"glottochronology/internal/app/redis"
	"glottochronology/internal/app/repository"
)

// @title       Glotto API
// @version     1.0
// @description API для расчёта времени расхождения языков
// @host        localhost:8080
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization
func main() {
	_ = godotenv.Load()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	dbString := dsn.FromEnv()
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	rep, err := repository.New(repository.Settings{
		PostgresDSN:    dbString,
		MinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:    os.Getenv("MINIO_BUCKET"),
		MinioUseSSL:    useSSL,
		MinioPublicURL: os.Getenv("MINIO_PUBLIC_URL"),
	})
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	ctx := context.Background()
	rc, err := redis.New(ctx)
	if err != nil {
		logrus.Fatalf("error initializing redis: %v", err)
	}
	defer rc.Close()

	h := handler.NewHandler(rep, rc)
	h.RegisterStatic(router)
	h.RegisterHandler(router)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Application start!")
	if err := router.Run(":8080"); err != nil {
		logrus.Fatal(err)
	}
}
