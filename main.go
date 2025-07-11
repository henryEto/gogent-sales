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
	APIPort     string
)

var ToolFunctions = getCompletionTools()

func main() {
	err := loadEnv()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handlerGeneric)
	http.HandleFunc("/v1/chat/completions", chatCompletionsHandler)

	log.Printf("Server starting on port%s...\n", APIPort)
	log.Fatal(http.ListenAndServe(APIPort, nil))
}

func loadEnv() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	GeminiKey = os.Getenv("GEMINI_API_KEY")
	GeminiModel = os.Getenv("GEMINI_MODEL")
	APIPort = fmt.Sprintf(":%v", os.Getenv("API_PORT"))

	return nil
}
