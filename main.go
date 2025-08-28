// main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/krtu0p/code-reviewer/handlers"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	if os.Getenv("OPENROUTER_API_KEY") == "" {
		log.Fatal("OPENROUTER_API_KEY environment variable is not set")
	}

	r := gin.Default()
	r.POST("/review", handlers.ReviewHandler)

	log.Println("AI code reviewer listening on port :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}
