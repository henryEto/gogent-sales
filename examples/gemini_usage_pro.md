
Of course! Here is a comprehensive guide on how to use the Gemini API with Go,
 starting from the basics and moving to more advanced features like streaming,
 multimodal input, and chat.

### Prerequisites

1.  **Go Environment:** Make sure you have Go installed and configured on your
 system. You can check with `go version`.
2.  **Gemini API Key:** You need an API key to access the Gemini API.
    *   Go to [Google AI Studio](https://aistudio.google.com/).
    *   Sign in with your Google account.
    *   Click on **"Get API key"** and create a new key.
    *   **Important:** For security, it's best to set your API key as an envir
onment variable rather than hardcoding it in your application.

    ```bash
    # In your terminal (for Linux/macOS)
    export GEMINI_API_KEY="YOUR_API_KEY_HERE"

    # In PowerShell (for Windows)
    $env:GEMINI_API_KEY="YOUR_API_KEY_HERE"
    ```

### Step 1: Project Setup and Installation

First, create a new directory for your project and initialize a Go module.

```bash
mkdir go-gemini-example
cd go-gemini-example
go mod init example.com/gemini
```

Next, install the official Google Generative AI SDK for Go:

```bash
go get github.com/google/generative-ai-go/genai
```

This will add the necessary dependency to your `go.mod` and `go.sum` files.

---

### Step 2: Basic Text Generation (GenerateContent)

This is the simplest use case: sending a text prompt and getting a text respon
se.

Create a file named `main.go`:

```go
package main

import (
        "context"
        "fmt"
        "log"
        "os"

        "github.comcom/google/generative-ai-go/genai"
        "google.golang.org/api/option"
)

func main() {
        // 1. Get the API Key from the environment variable
        apiKey := os.Getenv("GEMINI_API_KEY")
        if apiKey == "" {
                log.Fatal("GEMINI_API_KEY environment variable not set.")
        }

        // 2. Create a new client
        ctx := context.Background()
        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatalf("Failed to create client: %v", err)
        }
        defer client.Close()

        // 3. Select the model
        // See https://ai.google.dev/models/gemini for a list of available mod
els
        model := client.GenerativeModel("gemini-1.5-flash-latest")

        // 4. Send the prompt and get the response
        prompt := genai.Text("Write a short, fun story about a gopher who lear
ns to code in Go.")
        resp, err := model.GenerateContent(ctx, prompt)
        if err != nil {
                log.Fatalf("Failed to generate content: %v", err)
        }

        // 5. Print the response
        printResponse(resp)
}

// printResponse iterates through the parts of the response and prints the tex
t.
func printResponse(resp *genai.GenerateContentResponse) {
        for _, cand := range resp.Candidates {
                if cand.Content != nil {
                        for _, part := range cand.Content.Parts {
                                fmt.Println(part)
                        }
                }
        }
        fmt.Println("---")
}
```

**To run it:**

```bash
go run main.go
```

---

### Step 3: Streaming Responses

For longer responses, streaming provides a better user experience by showing t
he text as it's being generated.

The key change is using `GenerateContentStream` instead of `GenerateContent`.

```go
// Replace the call to model.GenerateContent with this:

fmt.Println("--- Streaming Response ---")

// Use GenerateContentStream for streaming
iter := model.GenerateContentStream(ctx, prompt)

for {
    resp, err := iter.Next()
    if err == iterator.Done {
        break
    }
    if err != nil {
        log.Fatalf("Streaming iteration failed: %v", err)
    }

    // In a stream, we print each part as it arrives.
    for _, cand := range resp.Candidates {
        for _, part := range cand.Content.Parts {
            // Use fmt.Print to see the text build up on one line
            fmt.Print(part)
        }
    }
}
fmt.Println("\n--- End of Stream ---")
```

You will need to add `"google.golang.org/api/iterator"` to your imports.

---

### Step 4: Multimodal Input (Text and Image)

Gemini models can understand both text and images in the same prompt. For this
, you need a multimodal model like `gemini-1.5-pro-latest`.

1.  Save an image (e.g., `gopher.jpg`) in your project directory.
2.  Update your `main.go` to read the image and include it in the prompt.

```go
package main

import (
        "context"
        "fmt"
        "log"
        "mime"
        "os"
        "path/filepath"

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/option"
)

func main() {
        // ... (Client setup is the same as before) ...
        apiKey := os.Getenv("GEMINI_API_KEY")
        // ...

        ctx := context.Background()
        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatal(err)
        }
        defer client.Close()

        // Use a model that supports multimodal input
        model := client.GenerativeModel("gemini-1.5-pro-latest")

        // 1. Read the image data
        imagePath := "gopher.jpg" // Change this to your image file
        imgData, err := os.ReadFile(imagePath)
        if err != nil {
                log.Fatalf("Failed to read image file: %v", err)
        }

        // 2. Create the multipart prompt
        prompt := []genai.Part{
                genai.ImageData(mime.TypeByExtension(filepath.Ext(imagePath)),
 imgData),
                genai.Text("What is in this image? Be creative in your descrip
tion."),
        }

        // 3. Generate the content
        resp, err := model.GenerateContent(ctx, prompt...)
        if err != nil {
                log.Fatalf("Failed to generate content: %v", err)
        }

        printResponse(resp)
}

func printResponse(resp *genai.GenerateContentResponse) {
        // ... (same as before) ...
}
```

---

### Step 5: Chat Sessions (Conversational History)

To build a chatbot, you need the model to remember the context of the conversa
tion. The SDK manages this easily with a `ChatSession`.

```go
package main

import (
        "bufio"
        "context"
        "fmt"
        "log"
        "os"

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/iterator"
        "google.golang.org/api/option"
)

func main() {
        // ... (Client setup is the same as before) ...
        apiKey := os.Getenv("GEMINI_API_KEY")
        // ...

        ctx := context.Background()
        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatal(err)
        }
        defer client.Close()

        model := client.GenerativeModel("gemini-1.5-flash-latest")

        // Start a chat session
        cs := model.StartChat()

        // Use a scanner to read user input from the console
        scanner := bufio.NewScanner(os.Stdin)
        fmt.Println("Chat started. Type 'quit' to exit.")
        fmt.Print("You: ")

        for scanner.Scan() {
                userInput := scanner.Text()
                if userInput == "quit" {
                        break
                }

                fmt.Println("Gemini:")
                // Send the user message and stream the response
                iter := cs.SendMessageStream(ctx, genai.Text(userInput))
                for {
                        resp, err := iter.Next()
                        if err == iterator.Done {
                                break
                        }
                        if err != nil {
                                log.Fatal(err)
                        }
                        // Print each part of the response
                        for _, cand := range resp.Candidates {
                                for _, part := range cand.Content.Parts {
                                        fmt.Print(part)
                                }
                        }
                }
                fmt.Println("\n--------------------")
                fmt.Print("You: ")
        }

        if err := scanner.Err(); err != nil {
                log.Fatalf("Error reading input: %v", err)
        }
}
```

In this example, the `ChatSession` (`cs`) automatically keeps track of the con
versation history. Each call to `cs.SendMessageStream` adds the new user messa
ge and the model's response to the history for the next turn.

---

### Additional Concepts

#### Configuring the Model (`GenerationConfig`)

You can control the model's behavior by setting parameters like temperature (r
andomness) and max output tokens.

```go
model := client.GenerativeModel("gemini-1.5-flash-latest")
model.GenerationConfig = &genai.GenerationConfig{
    Temperature:     0.9,     // Higher is more creative (0.0 - 1.0)
    TopK:            1,
    TopP:            1,
    MaxOutputTokens: 2048,
}
```

#### Error Handling

It's crucial to handle potential errors, especially those related to safety se
ttings. A response might be blocked if the prompt or the generated content vio
lates Google's safety policies.

```go
resp, err := model.GenerateContent(ctx, prompt)
if err != nil {
    log.Fatalf("Failed to generate content: %v", err)
}

// Check if the response was blocked
if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
    log.Println("Response was empty, potentially blocked.")
    // You can inspect PromptFeedback for details
    if resp.PromptFeedback != nil {
        log.Printf("Prompt Feedback: Blocked due to %s", resp.PromptFeedback.B
lockReason)
    }
    return
}
```

 󰊠                                  enrique   /Projects/copo-agent took 46s 
