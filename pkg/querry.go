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

func SearchGemini(query string) (string, error) {

	prompt := `
	you are a good assistant in answering.
	the user asked about this question :
	%s
	answer this in 5 sentense
	
	`

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("error on load config", err)
		return "", err
	}

	apiKey := cfg.GeminiApiKey // Use Gemini API Key from config
	if apiKey == "" {
		return "", fmt.Errorf("API key is missing")
	}

	pro := fmt.Sprintf(prompt, query)
	requestBody := models.GeminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: pro},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("JSON Marshal error:", err)
		return "", err
	}

	req, err := http.NewRequest("POST", geminiURL+"key="+apiKey, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("req", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

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

	var responseBody models.GeminiResponse
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		fmt.Println("JSON Unmarshal error:", err)
		return "", err
	}

	// Extract response content
	if len(responseBody.Candidates) > 0 && len(responseBody.Candidates[0].Content.Parts) > 0 {
		return responseBody.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("empty response from Gemini API")
}
