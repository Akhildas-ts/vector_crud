package pkg

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/pinecone"
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

		log.Println("Received Query:", req.Query)

		// Convert the query string into a vector
		queryVector, err := ConvertQueryToVector(client, req.Query, ctx) // Replace with actual conversion logic
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert query to vector"})
			return
		}

		log.Println("Generated Query Vector:", queryVector)

		// Ensure queryVector has valid values
		if len(queryVector) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Generated query vector is empty"})
			return
		}

		// Fetch the top 5 most related vectors from Pinecone
		res, err := idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
			Vector:          queryVector,
			TopK:            5, // Fetch top 5 matches
			IncludeValues:   true,
			IncludeMetadata: true,
		})
		if err != nil {
			log.Printf("Failed to fetch vectors: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vectors"})
			return
		}

		// Log the response for debugging
		log.Printf("Query Results: %+v", res)

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

		// Return the structured response as JSON
		c.JSON(http.StatusOK, gin.H{"results": responseData})
	}
}

func ConvertQueryToVector(client *pinecone.Client, query string, ctx context.Context) ([]float32, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	embeddingModel := "multilingual-e5-large"

	// Define embedding parameters
	docParameters := pinecone.EmbedParameters{
		InputType: "passage",
		Truncate:  "END",
	}

	// Convert query string into vector format by calling the embedding service
	docEmbeddingsResponse, err := client.Inference.Embed(ctx, &pinecone.EmbedRequest{
		Model:      embeddingModel,
		TextInputs: []string{query}, // Single query string
		Parameters: docParameters,
	})

	if err != nil {
		log.Fatalf("Failed to embed query: %v", err)
		return nil, fmt.Errorf("error is %v", err)
	}

	// Assuming that the response contains a single vector for the query (index 0)
	embedding := (*docEmbeddingsResponse.Data)[0]
	values := embedding.Values

	// Return the vector as a slice of float32
	return *values, nil
}

// func SearchHandler(idxConnection *pinecone.IndexConnection) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx := context.Background()

// 		// Define a struct to bind the JSON body
// 		type SearchRequest struct {
// 			VectorIDs []string `json:"ids"` // Expecting a list of vector IDs
// 		}

// 		var req SearchRequest
// 		// Bind JSON body to the struct
// 		if err := c.ShouldBindJSON(&req); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
// 			return
// 		}

// 		// Check if any vector IDs were provided
// 		if len(req.VectorIDs) == 0 {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "No vector IDs provided"})
// 			return
// 		}

// 		// Fetch vectors from Pinecone
// 		res, err := idxConnection.FetchVectors(ctx, req.VectorIDs)

// 		fmt.Println("Requested Vector IDs:", req.VectorIDs)

// 		if err != nil {
// 			log.Printf("Failed to fetch vectors: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vectors"})
// 			return
// 		}

// 		// Format the response to show relevant data (ID, Metadata)
// 		var responseData []map[string]interface{}

// 		for _, vector := range res.Vectors {
// 			data := map[string]interface{}{
// 				"ID":     vector.Id,
// 				"Fields": vector.Metadata, // Assuming metadata has the 'text' field
// 			}
// 			responseData = append(responseData, data)
// 		}

// 		// Return the structured response as JSON
// 		c.JSON(http.StatusOK, responseData)
// 	}
// }
