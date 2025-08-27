package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"./handlers"
)

func main() {

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		log.Fatal("DEEPSEEK_API_KEY environment variable is not set")
	}

	r := gin.Default()
	r.POST("/review", handlers.ReviewHandler)

	log.Println("AI code reviewer EINO deepsek port :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}

}
