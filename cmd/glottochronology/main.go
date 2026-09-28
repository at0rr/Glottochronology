package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"glottochronology/internal/app/dsn"
	"glottochronology/internal/app/handler"
	"glottochronology/internal/app/repository"
)

func main() {
	_ = godotenv.Load()
	err := godotenv.Load()
	if err != nil {
		log.Println("godotenv error:", err)
	}
	fmt.Println("DB_HOST =", os.Getenv("DB_HOST"))

	router := gin.Default()

	dbString := dsn.FromEnv()
	fmt.Println("DSN:", dbString)

	rep, err := repository.New(dbString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep)
	hand.RegisterStatic(router)
	hand.RegisterHandler(router)

	log.Println("Application start!")
	if err := router.Run(":8080"); err != nil {
		logrus.Fatal(err)
	}
	log.Println("Application terminated!")
}
