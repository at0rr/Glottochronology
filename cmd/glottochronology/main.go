package main

import (
	"log"
	"glottochronology/internal/api" // путь к пакету api
)

func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
