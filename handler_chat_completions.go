package main

import (
	"context"
	"copo-ai-agent/internal/database"
	"copo-ai-agent/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"google.golang.org/genai"
)

// Define structs to match OpenAI Chat Completions API request/response for simplicity

type OpenAIRequest struct {
	Messages []OpenAIMessage `json:"messages"`
	Model    string          `json:"model"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
}

type OpenAIChoice struct {
	Index   int           `json:"index"`
	Message OpenAIMessage `json:"message"`
}

type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func chatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	// Check for correct method POST
	if r.Method != http.MethodPost {
		log.Println("method not allowed...")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode request
	var req OpenAIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("invalid request body...")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract the latest user message
	var userQuery string
	if len(req.Messages) > 0 {
		userQuery = req.Messages[len(req.Messages)-1].Content
	} else {
		log.Println("no messages in request...")
		http.Error(w, "No messages in request", http.StatusBadRequest)
		return
	}

	// Process suer query
	geminiResponseContent, err := processUserQuery(userQuery)
	if err != nil {
		log.Fatalf("failed to process user query: %v", err)
		http.Error(w, "Failed to get response from gemini", http.StatusInternalServerError)
		return
	}

	openAIResp := OpenAIResponse{
		ID:      "chatcmpl-custom-" + uuid.New().String(), // You'll need a UUID generator
		Object:  "chat.completion",
		Created: 0, // Placeholder
		Model:   "gemini-pro",
		Choices: []OpenAIChoice{
			{
				Index: 0,
				Message: OpenAIMessage{
					Role:    "assistant",
					Content: geminiResponseContent,
				},
			},
		},
		Usage: OpenAIUsage{
			PromptTokens:     0, // Not provided by Gemini directly in this format
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResp)
}

func processUserQuery(userQuery string) (string, error) {
	// Create context and client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: GeminiKey})
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	listaProductosFunc := genai.FunctionDeclaration{
		Name:        "obtenerListaDeProductos",
		Description: "Obtiene una lista de todos los productos disponibles en la base de datos con información básica de código, descripción, línea, sublínea, marca, y score de popularidad y la devuelve como una cadena JSON.",
		Parameters:  &genai.Schema{Type: genai.TypeObject},
		Response:    &genai.Schema{Type: genai.TypeString},
	}

	chat, err := client.Chats.Create(
		ctx,
		GeminiModel,
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: "Eres un agente asistente con una serie de herramientas específicas. Si tus herramientas no son suficientes para contestar al usuario, dícelo de forma amigable y hazle saber tus capacidades. Tus respuestas deben ser concisas y amigables. Debes formatear la respuesta para ser utilizada directamente en un chat de WhatsApp"}},
			},
			Tools: []*genai.Tool{
				{
					FunctionDeclarations: []*genai.FunctionDeclaration{
						&listaProductosFunc,
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat: %w", err)
	}

	resp, err := chat.SendMessage(ctx, genai.Part{Text: userQuery})
	if err != nil {
		return "", fmt.Errorf("failed to send message %w", err)
	}

	part := resp.Candidates[0].Content.Parts[0]
	fc := part.FunctionCall
	if fc != nil {
		log.Printf("gemini wants to call function: %s\n", fc.Name)

		if fc.Name == "obtenerListaDeProductos" {
			result := obtenerListaDeProductos()
			log.Println("executed local function...")
			log.Println("sending function result back to Gemini...")
			prompt := fmt.Sprintf(
				"Eres un asistente. Se te ha dado el resultado de la lista de productos "+
					"que vendemos. Usa esta información para responder la pregunta del usuario. "+
					"Resultado: %s",
				result,
			)

			resp, err = chat.SendMessage(
				ctx,
				genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						Name: "add",
						Response: map[string]any{
							"result": result,
						},
					},
				},
				genai.Part{Text: prompt},
			)
			if err != nil {
				return "", fmt.Errorf("failed to send function response %w", err)
			}
		}
	}
	return resp.Text(), nil
}

func obtenerListaDeProductos() string {
	db, err := sql.Open("mysql", utils.GetConnString())
	if err != nil {
		return "ocurrió un error al obtener la lista de productos"
	}
	defer db.Close()
	queries := database.New(db)

	productos, err := queries.GetListOfProducts(context.Background())
	if err != nil {
		return "ocurrió un error al obtener la lista de productos"
	}

	jsonData, err := json.Marshal(productos)
	if err != nil {
		return "ocurrió un error al obtener la lista de productos"
	}

	return string(jsonData)
}
