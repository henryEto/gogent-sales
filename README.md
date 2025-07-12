# COPO AI Sales Agent

This project implements an internal AI agent designed to assist the sales team in generating quick, accurate, and context-aware responses to customer questions. It leverages a Go backend, Google's Gemini Large Language Model (LLM) for AI capabilities, and Open WebUI as the frontend interface. Real-time product data is retrieved from a local MariaDB database.

## Project Goal

The primary objective is to streamline the sales process by providing an AI assistant that can instantly answer customer queries about products, including descriptions, presentations, average weights, prices, and stock information, using live data.

## Technology Stack

* **Backend API & Logic:** Go (Golang)
* **Large Language Model (LLM):** Google Gemini API
* **Database:** MariaDB (existing local database)
* **Frontend User Interface:** Open WebUI (running as a Podman container)

## Architecture Overview

The Go backend acts as the central intelligence, orchestrating data retrieval from the MariaDB database and interacting with the Gemini LLM. It exposes an OpenAI Chat Completions API-compatible endpoint, allowing seamless integration with Open WebUI.

```mermaid
graph TD
    A[Open WebUI Frontend] -->|HTTP POST /v1/chat/completions| B(Go Backend API)
    B -->|Gemini API Call| C(Google Gemini LLM)
    B -->|Database Queries (sqlc)| D(MariaDB Database)
    C --|Function Calls / Responses| B
    D --|Product Data| B
    B -->|OpenAI-compatible Response| A
````

## Backend (Go) - Core Components

The Go backend is responsible for:

  * **HTTP API Endpoint:** Exposes a `/v1/chat/completions` endpoint compatible with the OpenAI Chat Completions API for integration with Open WebUI.
  * **Database Interaction Layer:** Connects to the MariaDB database and performs targeted data retrieval using `database/sql` and `sqlc` generated code.
  * **Gemini API Integration:** Utilizes the Google Go SDK for the Gemini API to send prompts and receive generated responses.
  * **Prompt Engineering & Orchestration Logic:** Analyzes user queries, dynamically calls appropriate database functions, constructs comprehensive prompts for the Gemini LLM by combining the query with retrieved data, and parses Gemini's responses.
  * **Tool Usage:** Implements various tools (Go functions) that the Gemini LLM can call to retrieve specific product information from the database (e.g., product lists, product details by search term, brand, category, or code).
  * **Response Formatting & Error Handling:** Formats the LLM's response into the expected OpenAI-compatible format and includes robust error handling.

## Frontend (Open WebUI) - Interaction

Open WebUI provides the chat interface for the sales team. It is configured to:

  * Point to the Go agent's OpenAI-compatible API endpoint.
  * Allow sales team members to type customer questions.
  * Send these questions to the Go backend and display the AI-generated responses.

## Getting Started

### Prerequisites

  * Go (Golang) installed
  * Podman (or Docker) installed
  * Access to a MariaDB database with your product data
  * Google Gemini API Key

### Setup

1.  **Clone the repository:**

    ```bash
    git clone <repository_url>
    cd copo-ai-agent
    ```

2.  **Environment Variables:**
    Create a `.env` file in the root directory of the project with the following variables:

    ```env
    GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
    GEMINI_MODEL="gemini-pro" # Or your preferred Gemini model
    API_PORT="8504" # Or your desired port for the Go backend
    DB_USER="your_db_user"
    DB_PASSWORD="your_db_password"
    DB_HOST="your_db_host" # e.g., 127.0.0.1 or localhost if on the same machine
    DB_PORT="3306" # Or your MariaDB port
    DB_NAME="your_database_name"
    ```

3.  **Database Schema (Conceptual):**
    Ensure your MariaDB database has tables containing product information. The Go backend expects to query for product codes, details, prices, stock, etc. You will likely use `sqlc` to generate Go code for your specific database schema. An example of the data points expected by the tool functions are:

      * Product Codes
      * Description
      * Line
      * Subline
      * Brand
      * Stock (Existencia)
      * Popularity
      * Average Weight
      * Pieces per box
      * Tiered Pricing (Detalle, Medio Mayoreo, Mayoreo)

4.  **Run the Go Backend:**

    ```bash
    go run main.go handler_chat_completions.go handler_generic.go openai_structs.go completion_tools.go product_functions.go prompts.go
    ```

    The server will start on the port specified in your `.env` file (e.g., `http://localhost:8504`).

5.  **Run Open WebUI:**
    Use the provided `podman-compose.yml` (or `docker-compose.yml` if you adapt it for Docker) to run Open WebUI.

    ```bash
    podman-compose up -d
    ```

    This will start Open WebUI, accessible at `http://localhost:3001` (or the port you mapped).

6.  **Configure Open WebUI:**
    The `open-webui-config.json` file provides an example configuration.

      * Access Open WebUI in your browser (`http://localhost:3001`).
      * Go to **Settings** -\> **Connections**.
      * Add a new OpenAI connection:
          * **Enable:** True
          * **API Base URL:** `http://host.docker.internal:8504/v1` (Note: `host.docker.internal` is used to reach the host machine's Go backend from within the Podman container. If you're using Docker Desktop on Windows/macOS, it's also `host.docker.internal`. On Linux with Podman, you might need to find your host's IP address and use that instead, e.g., `http://192.168.1.X:8504/v1`).
          * **API Key:** `my_super_secure_key` (This is a placeholder as the Go backend doesn't currently validate this specific key, but Open WebUI requires it. You can put any string here.)
          * **Connection Type:** `local`
          * **Model IDs:** `COPO-AI` (This is the model name Open WebUI will display for your agent. It should match the `Model` field in the `OpenAIResponse` struct in `openai_structs.go`, which is set to `GeminiModel` in `main.go`. Ensure your `GEMINI_MODEL` environment variable is set accordingly, for example to `COPO-AI`.)

## Usage

Once both the Go backend and Open WebUI are running and configured:

1.  Open Open WebUI in your browser.
2.  Select the "COPO-AI" model (or whatever you named it in the Open WebUI configuration).
3.  Start typing questions related to your products in the chat interface. The AI agent will leverage the Gemini LLM and your database to provide informed responses.

## Project Structure

  * `main.go`: Entry point for the Go application, handles environment loading and HTTP server setup.
  * `handler_chat_completions.go`: Implements the OpenAI Chat Completions API compatible endpoint and orchestrates the Gemini LLM interaction and tool calls.
  * `handler_generic.go`: A generic HTTP handler for debugging and request logging.
  * `openai_structs.go`: Defines the Go structs for OpenAI Chat Completions API requests and responses.
  * `completion_tools.go`: Defines the `FunctionTool` struct and registers the available tools (`obtenerListaProductos`, `obtenerInformacionPorBusqueda`, `obtenerInformacionPorMarca`, `obtenerInformacionPorLineaSublinea`, `obtenerInformacionPorCodigo`) that Gemini can call.
  * `product_functions.go`: Contains the actual Go functions that interact with the database to retrieve product information, corresponding to the `FunctionTool` implementations.
  * `prompts.go`: Stores the system prompt used to guide the Gemini LLM's behavior and response formatting.
  * `internal/database/`: (Assumed) Directory for `sqlc`-generated database query code and database models.
  * `internal/utils/`: (Assumed) Directory for utility functions, e.g., `GetConnString()`.
  * `podman-compose.yml`: Configuration for running Open WebUI as a Podman container.
  * `open-webui-config.json`: Example configuration for Open WebUI.


## LICENSE
This project is a personal work and is provided for viewing purposes only.
All rights reserved. Unauthorized use, reproduction, or distribution
of this code, in whole or in part, is strictly prohibited.
