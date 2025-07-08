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
		Name: "obtenerListaDeProductos",
		Description: "Obtiene una lista de todos los productos disponibles " +
			"en la base de datos y la devuelve como una cadena JSON. " +
			"La lista contiene información básica: " +
			"código, descripción, línea, sublínea, marca y score de popularidad " +
			"Un score de popularidad alto indica que es un producto muy vendido o popular. " +
			"Esta lista se puede usar para: " +
			"saber que tipo de productos vendemos (lineas y sublineas), " +
			"saber que marcas de productos tenemos, " +
			"saber los códigos y nombres de productos que vendemos, " +
			"saber cuales son los productos más populares o vendidos (score de popularidad)",
		Parameters: &genai.Schema{Type: genai.TypeObject},
		Response:   &genai.Schema{Type: genai.TypeString},
	}

	infoProductosFunc := genai.FunctionDeclaration{
		Name:        "obtenerInformacionDeProductos",
		Description: "Obtiene información detallada para una lista de códigos de productos específicos. Devuelve un JSON con el código, descripción, línea, sublínea, marca, existencia, peso promedio por caja, piezas por caja, peso promedio por pieza, y diferentes escalas de precios (detalle, medio mayoreo, mayoreo, especial).",
		Parameters: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"productCodes": {
					Type:        genai.TypeArray,
					Description: "Una lista de códigos de productos (strings) para los que se desea obtener información.",
					Items:       &genai.Schema{Type: genai.TypeString},
				},
			},
			Required: []string{"productCodes"},
		},
		Response: &genai.Schema{Type: genai.TypeString},
	}

	chat, err := client.Chats.Create(
		ctx,
		GeminiModel,
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: "Eres un agente asistente de ventas. " +
					"Debes usar y procesar la información obtenida con estas " +
					"herramientas para tratar de responder lo mejor posible a las preguntas " +
					"del usuario. Tus respuestas deben ser concisas y amigables. " +
					"Debes formatear la respuesta para ser utilizada directamente en un chat de WhatsApp. " +
					"Si no hay forma de responder a la pregunta con las herramientas " +
					"que se te han brindado, entonces hazlo saber al usuario."}},
			},
			Tools: []*genai.Tool{
				{
					FunctionDeclarations: []*genai.FunctionDeclaration{
						&listaProductosFunc,
						&infoProductosFunc,
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

	db, err := sql.Open("mysql", utils.GetConnString())
	if err != nil {
		log.Printf("failed to open db (%s): %v", utils.GetConnString(), err)
		return "", fmt.Errorf("ocurrió un error al obtener la lista de productos")
	}
	defer db.Close()
	queries := database.New(db)

	for {
		if len(resp.FunctionCalls()) > 0 {
			// log.Println("found FunctionCall...")
			fc := resp.FunctionCalls()[0]

			// log.Printf("executing %s()...", fc.Name)
			var result string
			switch fc.Name {
			case "obtenerListaDeProductos":
				result = obtenerListaDeProductos(queries)
			case "obtenerInformacionDeProductos":
				var codigos []string
				if argCodes, ok := fc.Args["productCodes"]; ok {
					if codesSlice, ok := argCodes.([]any); ok {
						for _, v := range codesSlice {
							if str, ok := v.(string); ok {
								codigos = append(codigos, str)
							}
						}
					}
				} else {
					log.Println("failed to extract args from FunctionCall...")
				}
				result = obtenerInformacionDeProductos(queries, codigos)
			default:
				result = fmt.Sprintf("failed to call %s() function", fc.Name)
			}
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

	return resp.Text(), nil
}

func obtenerListaDeProductos(queries *database.Queries) string {
	productos, err := queries.GetListOfProducts(context.Background())
	if err != nil {
		log.Printf("failed to get products list: %v", err)
		return "ocurrió un error al obtener la lista de productos"
	}

	jsonData, err := json.Marshal(productos)
	if err != nil {
		log.Printf("failed to marshal results: %v", err)
		return "ocurrió un error al obtener la lista de productos"
	}

	return string(jsonData)
}

func obtenerInformacionDeProductos(queries *database.Queries, productCodes []string) string {
	infoProductos, err := queries.GetProductsInfo(context.Background(), productCodes)
	if err != nil {
		log.Printf("failed to get products info: %v", err)
		return "ocurrió un error al obtener la información de los productos"
	}

	jsonData, err := json.Marshal(infoProductos)
	if err != nil {
		log.Printf("failed to marshal results: %v", err)
		return "ocurrió un error al obtener la información de los productos"
	}

	return string(jsonData)
}
