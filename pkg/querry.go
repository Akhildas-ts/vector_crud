package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"vectorchat/pkg/config"
	"vectorchat/pkg/models"
)

const geminiURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?"

func SearchOpenAI(query string) (string, error) {
	prompt := fmt.Sprintf(`You are a highly skilled assistant in summarizing conversations.
The user provided an interview conversation spanning 5-6 pages.
%s
Summarize the conversation while including all key points, ensuring that no important details are missed.
Provide a clear and concise summary while maintaining the essence of the discussion.`, query)

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error on load config", err)
		return "", err
	}

	apiKey := cfg.OpenApiKey // Use OpenAI API Key from config
	if apiKey == "" {
		return "", fmt.Errorf("API key is missing")
	}

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

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("req", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("API response status:", resp.Status)
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("API response body:", string(body))
		return "", fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("body", err)
		return "", err
	}

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
