package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"vectorchat/pkg/config"
	"vectorchat/pkg/models"
)

const geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?"

func SearchOpenAI(query string) (string, error) {

	// Step 1: Create Prompt
	prompt := fmt.Sprintf(`You are a highly skilled assistant in summarizing conversations.
The user provided an interview conversation spanning 5-6 pages.
%s
Summarize the conversation while including all key points, ensuring that no important details are missed.
Provide a clear and concise summary while maintaining the essence of the discussion.`, query)

	// Step 2: Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error on load config", err)
		return "", err
	}

	// Step 3: Prepare API Request
	requestBody := models.OpenAIRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return "", err
	}

	// Step 4: Make API Request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Request creation error:", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenApiKey)

	apiStart := time.Now()
	resp, err := http.DefaultClient.Do(req)
	fmt.Println("Step 4 (API Request Time):", time.Since(apiStart))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Step 5: Read API Response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Body Read error:", err)
		return "", err
	}

	// Step 6: Parse API Response
	var responseBody models.OpenAIResponse
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return "", err
	}

	// Extract response content
	if len(responseBody.Choices) > 0 {

		return responseBody.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from OpenAI API")
}
