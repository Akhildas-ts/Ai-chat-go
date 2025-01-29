package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type Content struct {
	Parts []string `json:"Parts"`
	Role  string   `json:"Role"`
}

type Candidates struct {
	Content *Content `json:"Content"`
}

type ContentResponse struct {
	Candidates *[]Candidates `json:"Candidates"`
}

type RequestPayload struct {
	Question string `json:"question"`
}

type ResponsePayload struct {
	Response string `json:"response"`
}

func main() {
	// Load API key from .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API key not found in .env file")
	}

	// Initialize Gemini AI client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")

	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse request payload
		var payload RequestPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Ensure the question is provided
		if payload.Question == "" {
			http.Error(w, "Question is required", http.StatusBadRequest)
			return
		}

		// Generate content using Gemini AI
		prompt := []genai.Part{
			genai.Text(payload.Question),
		}
		resp, err := model.GenerateContent(ctx, prompt...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error generating response: %v", err), http.StatusInternalServerError)
			return
		}

		// Extract response from Gemini AI
		var generateResponse ContentResponse
		marshalResponse, _ := json.Marshal(resp)
		if err := json.Unmarshal(marshalResponse, &generateResponse); err != nil {
			http.Error(w, "Error parsing AI response", http.StatusInternalServerError)
			return
		}

		// Construct the response
		var aiResponse string
		for _, cad := range *generateResponse.Candidates {
			if cad.Content != nil {
				for _, part := range cad.Content.Parts {
					aiResponse += part
				}
			}
		}

		// Send the response back to the user
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ResponsePayload{Response: aiResponse})
	})

	// Start the HTTP server
	port := "8080"
	fmt.Printf("Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
