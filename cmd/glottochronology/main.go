package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"glottochronology/internal/app/dsn"
	"glottochronology/internal/app/handler"
	"glottochronology/internal/app/repository"
)

func main() {
	_ = godotenv.Load()

	router := gin.Default()

	dbString := dsn.FromEnv()
	fmt.Println("DSN:", dbString)

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

	h := handler.NewHandler(rep)
	h.RegisterStatic(router)
	h.RegisterHandler(router)

	log.Println("Application start!")
	if err := router.Run(":8080"); err != nil {
		logrus.Fatal(err)
	}
	log.Println("Application terminated!")
}
