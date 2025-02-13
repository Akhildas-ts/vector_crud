package pkg

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"vectorchat/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"github.com/sashabaranov/go-openai"
)

func SearchHandler(idxConnection *pinecone.IndexConnection, client *pinecone.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use context from request
		ctx := c.Request.Context()

		// Define a struct to bind the JSON body
		type SearchRequest struct {
			Query string `json:"query"` // Expecting a search query string
		}

		var req SearchRequest
		// Bind JSON body to the struct
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// Check if query string is provided
		if len(req.Query) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No query provided"})
			return
		}

		expandQuerry, err := ExpandQuery(req.Query)

		if err != nil {

			fmt.Println("error on expand querry")
			return
		}

		log.Println("expand querry Received :", expandQuerry)

		// Convert the query string into a vector
		queryVector, err := ConvertQueryToVector(client, expandQuerry, ctx) // Replace with actual conversion logic
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert query to vector"})
			return
		}

		// Ensure queryVector has valid values
		if len(queryVector) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Generated query vector is empty"})
			return
		}

		// Fetch the top 3 most related vectors from Pinecone
		res, err := idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
			Vector:          queryVector,
			TopK:            3, // Fetch top 5 matches
			IncludeValues:   true,
			IncludeMetadata: true,
		})
		if err != nil {
			log.Printf("Failed to fetch vectors: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vectors"})
			return
		}

		// Prepare the response as structured JSON
		var responseData []gin.H
		for _, match := range res.Matches {
			metadataData := ""

			if metadata := match.Vector.Metadata; metadata != nil {
				if dataField, exists := metadata.GetFields()["text"]; exists {
					metadataData = dataField.GetStringValue()
				}
			}

			responseData = append(responseData, gin.H{
				"id":    match.Vector.Id,
				"score": match.Score,
				"text":  metadataData,
			})
		}

		prompt := `
	The user asked: "%s"

	Expanded Query: "%s"

	Below are the retrieved search results:
	"%s"

	Using these results, provide a well-structured, relevant, and detailed response.
`
		finalPrompt := fmt.Sprintf(prompt, req.Query, expandQuerry, responseData)

		finalResp, err := SearchOpenAI(finalPrompt)

		if err != nil {

			c.JSON(http.StatusBadGateway, gin.H{"gemini resp": err})
			return
		}

		// Return the structured response as JSON
		c.JSON(http.StatusOK, gin.H{"results": finalResp})
	}
}

func ConvertQueryToVector(client *pinecone.Client, query string, ctx context.Context) ([]float32, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// embeddingModel := "multilingual-e5-large"

	// Define embedding parameters
	// docParameters := pinecone.EmbedParameters{
	// 	InputType: "passage",
	// 	Truncate:  "END",
	// }

	// start := time.Now()
	cfg, err := config.LoadConfig()
	if err != nil {

		return nil, err
	}
	openAiClient := openai.NewClient(cfg.OpenApiKey)
	embeddingRes, err := openAiClient.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: "text-embedding-ada-002",
		Input: query,
	})

	// Convert query string into vector format by calling the embedding service
	// docEmbeddingsResponse, err := client.Inference.Embed(ctx, &pinecone.EmbedRequest{
	// 	Model:      embeddingModel,
	// 	TextInputs: []string{query}, // Single query string
	// 	Parameters: docParameters,
	// })

	if err != nil {
		log.Fatalf("Failed to embed query: %v", err)
		return nil, fmt.Errorf("error is %v", err)
	}

	// Assuming that the response contains a single vector for the query (index 0)

	values := embeddingRes.Data[0].Embedding

	// Return the vector as a slice of float32
	return values, nil
}

func ExpandQuery(originalQuery string) (string, error) {
	// Call OpenAI (or any LLM) to generate an expanded query
	prompt := fmt.Sprintf(`Expand the following query by adding relevant keywords or synonyms while preserving intent: "%s"`, originalQuery)

	expandedQuery, err := SearchOpenAI(prompt)
	if err != nil {
		return "", err
	}

	return expandedQuery, nil
}
