package pkg

// get embedding

// func GetEmbedding(text string, apiKey string) ([]float32, error) {
// 	fmt.Println("Inside GetEmbedding function")

// 	url := env.EmbeddingAPIURL
// 	requestBody := models.EmbeddingRequest{
// 		Inputs: []string{text},
// 	}

// 	// Convert the request body to JSON
// 	jsonBody, err := json.Marshal(requestBody)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshaling request: %v", err)
// 	}

// 	// Create an HTTP request
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating request: %v", err)
// 	}

// 	// Set headers
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
// 	req.Header.Set("Content-Type", "application/json")

// 	// Make the HTTP request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("error making request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	// Read response body
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("error reading response: %v", err)
// 	}

// 	// Handle API errors
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
// 	}

// 	// Parse the JSON response
// 	var embedding models.EmbeddingResponse
// 	err = json.Unmarshal(body, &embedding)
// 	if err != nil {
// 		log.Println("Error decoding response:", err)
// 		return nil, fmt.Errorf("error decoding response: %v", err)
// 	}

// 	// Ensure embeddings exist
// 	if len(embedding) == 0 {
// 		return nil, fmt.Errorf("empty embedding response")
// 	}

// 	// Return the first embedding vector
// 	return embedding[0], nil
// }
