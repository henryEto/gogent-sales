
Using the Gemini API in Go is straightforward, thanks to the official Google G
enerative AI Go client library. This guide will walk you through the process, 
covering text generation, chat conversations, and vision capabilities.

---

### Prerequisites

1.  **Go Lang Installed:** Make sure you have Go 1.18 or higher installed on y
our system.
2.  **Google Cloud Project:**
    *   Go to the [Google Cloud Console](https://console.cloud.google.com/).
    *   Create a new project or select an existing one.
3.  **Enable the Gemini API:**
    *   In your Google Cloud project, navigate to **APIs & Services > Library*
*.
    *   Search for "Generative Language API" or "Gemini API" and enable it.
4.  **Create an API Key:**
    *   In your Google Cloud project, go to **APIs & Services > Credentials**.
    *   Click "CREATE CREDENTIALS" and choose "API Key".
    *   Copy the generated API key. **Keep this key secure and do not expose i
t in your code directly.** We'll use an environment variable.

---

### Step 1: Set up your Go Project

Create a new directory for your project and initialize a Go module:

```bash
mkdir gemini-go-example
cd gemini-go-example
go mod init gemini-go-example
```

---

### Step 2: Install the Go Client Library

Install the official Google Generative AI Go client library:

```bash
go get github.com/google/generative-ai-go/genai
```

---

### Step 3: Authenticate with your API Key

The easiest way for development is to set your API key as an environment varia
ble.

**Linux/macOS:**
```bash
export GOOGLE_API_KEY="YOUR_API_KEY_HERE"
```

**Windows (Command Prompt):**
```cmd
set GOOGLE_API_KEY="YOUR_API_KEY_HERE"
```

**Windows (PowerShell):**
```powershell
$env:GOOGLE_API_KEY="YOUR_API_KEY_HERE"
```

Replace `YOUR_API_KEY_HERE` with the actual API key you generated.

---

### Step 4: Basic Text Generation

Let's start with a simple example to generate text.

Create a file named `main.go`:

```go
package main

import (
        "context"
        "fmt"
        "log"
        "os"

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/option"
)

func main() {
        // 1. Get API Key from environment variable
        apiKey := os.Getenv("GOOGLE_API_KEY")
        if apiKey == "" {
                log.Fatal("GOOGLE_API_KEY environment variable not set")
        }

        // 2. Create a new context
        ctx := context.Background()

        // 3. Create a Gemini client
        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatal(err)
        }
        defer client.Close() // Ensure the client is closed when done

        // 4. Select the model (gemini-pro is for text-only tasks)
        model := client.GenerativeModel("gemini-pro")

        // 5. Generate content
        prompt := "Tell me a short, interesting fact about the ocean."
        resp, err := model.GenerateContent(ctx, genai.Text(prompt))
        if err != nil {
                log.Fatal(err)
        }

        // 6. Process the response
        fmt.Println("--- Text Generation ---")
        if len(resp.Candidates) > 0 {
                candidate := resp.Candidates[0]
                if len(candidate.Content.Parts) > 0 {
                        for _, part := range candidate.Content.Parts {
                                if text, ok := part.(genai.Text); ok {
                                        fmt.Println(string(text))
                                }
                        }
                } else {
                        fmt.Println("No content parts in the first candidate."
)
                }
        } else {
                fmt.Println("No candidates generated.")
                // Check for block reasons if no candidates
                if resp.PromptFeedback != nil && len(resp.PromptFeedback.Safet
yRatings) > 0 {
                        fmt.Println("Prompt blocked due to safety reasons:")
                        for _, rating := range resp.PromptFeedback.SafetyRatin
gs {
                                fmt.Printf("  Category: %s, Probability: %s\n"
, rating.Category, rating.Probability)
                        }
                }
        }
}

```

To run this:

```bash
go run main.go
```

You should see an interesting fact about the ocean printed to your console.

---

### Step 5: Chat Conversations (Multi-Turn)

The `gemini-pro` model can also handle multi-turn conversations. The client li
brary provides a `ChatSession` for this.

Add the following function to your `main.go` file (or create a new one):

```go
package main

import (
        "bufio"
        "context"
        "fmt"
        "log"
        "os"
        "strings"

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/option"
)

// ... (main function from above, or remove it for this example)

func main() {
        // ... (API key and client setup as before)
        apiKey := os.Getenv("GOOGLE_API_KEY")
        if apiKey == "" {
                log.Fatal("GOOGLE_API_KEY environment variable not set")
        }

        ctx := context.Background()
        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatal(err)
        }
        defer client.Close()

        // Start a chat session
        model := client.GenerativeModel("gemini-pro")
        cs := model.StartChat()

        fmt.Println("\n--- Chat Conversation (Type 'exit' to quit) ---")
        reader := bufio.NewReader(os.Stdin)

        for {
                fmt.Print("You: ")
                input, _ := reader.ReadString('\n')
                input = strings.TrimSpace(input)

                if strings.ToLower(input) == "exit" {
                        break
                }

                // Send message to the chat session
                resp, err := cs.SendMessage(ctx, genai.Text(input))
                if err != nil {
                        log.Printf("Error sending message: %v", err)
                        continue
                }

                // Process response
                if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.
Parts) > 0 {
                        for _, part := range resp.Candidates[0].Content.Parts 
{
                                if text, ok := part.(genai.Text); ok {
                                        fmt.Printf("Gemini: %s\n", string(text
))
                                }
                        }
                } else {
                        fmt.Println("Gemini: I didn't get a clear response. Ca
n you rephrase?")
                        if resp.PromptFeedback != nil && len(resp.PromptFeedba
ck.SafetyRatings) > 0 {
                                fmt.Println("  (Prompt may have been blocked d
ue to safety reasons)")
                        }
                }
        }

        fmt.Println("Chat ended.")
}

```

Run this:

```bash
go run main.go
```

Now you can type messages, and Gemini will respond, remembering the context of
 your conversation.

---

### Step 6: Vision Capabilities (Image Input)

The `gemini-pro-vision` model can accept both text and image input.

First, you'll need an image file in the same directory as your `main.go` (e.g.
, `image.jpg`).

```go
package main

import (
        "context"
        "fmt"
        "log"
        "os"
        "mime" // Required for detecting MIME type

        "github.com/google/generative-ai-go/genai"
        "google.golang.org/api/option"
)

// ... (main function from above, or remove it for this example)

func main() {
        // ... (API key and client setup as before)
        apiKey := os.Getenv("GOOGLE_API_KEY")
        if apiKey == "" {
                log.Fatal("GOOGLE_API_KEY environment variable not set")
        }

        ctx := context.Background()
        client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
        if err != nil {
                log.Fatal(err)
        }
        defer client.Close()

        // Select the vision model
        model := client.GenerativeModel("gemini-pro-vision")

        fmt.Println("\n--- Vision Capabilities ---")

        // Path to your image file
        imagePath := "example.jpg" // Make sure this image file exists in your
 project directory!

        // Determine MIME type based on file extension
        mimeType := mime.TypeByExtension(".jpg") // or ".png", ".jpeg", etc.
        if mimeType == "" {
                log.Fatalf("Could not determine MIME type for %s", imagePath)
        }

        // Create content parts: text and image
        // genai.FileData is for local files, genai.ImageData for base64 encod
ed
        promptParts := []genai.Part{
                genai.Text("What do you see in this picture? Describe it in de
tail."),
                genai.FileData(imagePath, mimeType),
        }

        // Generate content with image
        resp, err := model.GenerateContent(ctx, promptParts...)
        if err != nil {
                log.Fatal(err)
        }

        // Process the response
        if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) >
 0 {
                for _, part := range resp.Candidates[0].Content.Parts {
                        if text, ok := part.(genai.Text); ok {
                                fmt.Println(string(text))
                        }
                }
        } else {
                fmt.Println("No response from vision model.")
                if resp.PromptFeedback != nil && len(resp.PromptFeedback.Safet
yRatings) > 0 {
                        fmt.Println("  (Prompt may have been blocked due to sa
fety reasons)")
                }
        }
}

```

**Important:** Before running, place an image file (e.g., `example.jpg`) in yo
ur `gemini-go-example` directory.

Run this:

```bash
go run main.go
```

Gemini will analyze the image and provide a textual description.

---

### Important Considerations

*   **Error Handling:** Always check for `err != nil` after API calls. The `lo
g.Fatal` approach is fine for simple scripts, but in production, you'd want mo
re sophisticated error handling (e.g., logging, retries, user feedback).
*   **Context:** `context.Background()` is used for simplicity. In larger appl
ications, you'd pass a context through function calls, potentially with timeou
ts (`context.WithTimeout`) or cancellation (`context.WithCancel`).
*   **Streaming:** For longer responses, `model.GenerateContentStream` can be 
used to receive parts of the response as they are generated, improving user ex
perience.
*   **Safety Settings:** You can configure safety settings to control content 
generation. See `model.SetSafetySettings`.
*   **Generation Configuration:** You can adjust parameters like `temperature`
 (randomness), `maxOutputTokens`, and `topP`/`topK` using `model.SetGenerative
Config`.
*   **Cost:** Be aware of the API usage costs. Monitor your usage in the Googl
e Cloud Console.
*   **Rate Limiting:** Gemini API has rate limits. If you make too many reques
ts too quickly, you might receive errors. Implement exponential backoff for re
tries if building a robust application.
*   **Authentication for Production:** For production applications, using Serv
ice Accounts and Google Cloud IAM roles is more secure and robust than directl
y using API keys from environment variables. You'd typically use `option.WithC
redentialsFile("path/to/your/service-account-key.json")`.

This comprehensive guide should give you a solid foundation for building appli
cations with the Gemini API in Go!

 󰊠                                  enrique   /Projects/copo-agent took 28s 
