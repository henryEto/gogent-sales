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

func chatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	// Check for correct method POST
	if r.Method != http.MethodPost {
		log.Printf("method not allowed: %v...\n", r.Method)
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
		log.Printf("failed to process user query: %v\n", err)
		http.Error(w, "Failed to get response from gemini", http.StatusInternalServerError)
		return
	}

	// Generate OpenAIResponse struct
	openAIResp := OpenAIResponse{
		ID:      "chatcmpl-custom-" + uuid.New().String(),
		Object:  "chat.completion",
		Created: 0,
		Model:   GeminiModel,
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
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(openAIResp)
}

func processUserQuery(userQuery string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: GeminiKey})
	if err != nil {
		return "", fmt.Errorf("failed to create client: %w", err)
	}

	chat, err := client.Chats.Create(
		ctx,
		GeminiModel,
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: getSystemPrompt()}},
			},
			Tools: []*genai.Tool{
				{
					FunctionDeclarations: ToolFunctions.getDeclarationsList(),
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

	db, err := sql.Open("mysql", utils.GetConnString())
	if err != nil {
		log.Printf("failed to open db (%s): %v", utils.GetConnString(), err)
		return "", fmt.Errorf("ocurrió un error al obtener la lista de productos")
	}
	defer db.Close()
	queries := database.New(db)

	for {
		log.Printf("total usage: %v tokens\n", resp.UsageMetadata.TotalTokenCount)

		if len(resp.FunctionCalls()) > 0 {
			// log.Println("found FunctionCall...")
			fc := resp.FunctionCalls()[0]

			// log.Printf("executing %s()...", fc.Name)
			var result string

			functionTool := ToolFunctions.getToolByName(fc.Name)
			result = functionTool.Function(queries, fc.Args)
			// log.Println("sending function result back to Gemini...")
			resp, err = chat.SendMessage(
				ctx,
				genai.Part{
					FunctionResponse: &genai.FunctionResponse{
						Name: "obtenerListaDeProductos",
						Response: map[string]any{
							"result": result,
						},
					},
				},
			)
			if err != nil {
				log.Printf("failed to send function response: %v", err)
				return "", fmt.Errorf("failed to send function response %w", err)
			}
		} else {
			// log.Println("no FunctionCall found...")
			break
		}
	}

	return formatResponse(resp.Text()), nil
}

func formatResponse(response string) string {
	header := `*¡Hola! 😊 Gracias por tu interés en nuestros productos!*

🚚 Hacemos entregas en Tula, Tepeji, Chapantongo, Jilotepec, Huehuetoca, Ixmiquilpan, Mixquiahuala y alrededores.`
	foot := `📍 También puedes visitarnos aquí: https://maps.app.goo.gl/QDv4HnqqJhqQ24BP8?g_st=ac
📲 Mándanos mensaje por WhatsApp: https://wa.me/527731819900
🐔 *COPOCAR* agradece tu preferencia!🙏`

	return fmt.Sprintf("%s\n\n\n%s\n\n\n%s", header, response, foot)
}
