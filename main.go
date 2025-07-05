package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// For API key

	"github.com/joho/godotenv"
)

var (
	GeminiKey   string
	GeminiModel string
)

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handlerGeneric)
	http.HandleFunc("/v1/chat/completions", chatCompletionsHandler)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	GeminiKey = os.Getenv("GEMINI_API_KEY")
	GeminiModel = os.Getenv("GEMINI_MODEL")

	return nil
}
