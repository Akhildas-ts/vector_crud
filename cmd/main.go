package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"vectorchat/pkg"
	"vectorchat/pkg/config"
	"vectorchat/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"github.com/joho/godotenv"
	
)

func main() {



	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
	cfg, err := config.LoadConfig()

	if err != nil {
		fmt.Println("err", err)
		fmt.Errorf("load config error ", err)

		return
	}
	r := gin.Default()

	fmt.Println("config.Env ", cfg.ApiKey)
	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: cfg.ApiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create Client: %v", err)
	}

	ctx := context.Background()

	r.POST("/upsert", func(c *gin.Context) {
		request := []models.EmbeddingRequest{}

		if err := c.ShouldBindJSON(&request); err != nil {

			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		err := pkg.Upsert(pc, request, ctx)

		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "upsert issue "})
			return
		}

	})

	idxConnection, err := pc.Index(pinecone.NewIndexConnParams{Host: cfg.PineconeHost, Namespace: "example-namespace"})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host: %v", err)
	}

	r.POST("/search", pkg.SearchHandler(idxConnection, pc, ctx))

	log.Println("Starting server on port 8080...")
	r.Run(":8080")
}

// r.POST("/search", func(c *gin.Context) {
// Start the server on port 8080

// // Convert text to vector using embedding API
// vectorValues, err := pkg.GetEmbedding(request.Inputs[0], env.EmbeddingAPIKey)
// if err != nil {

// 	fmt.Println("error is ", err)
// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate embeddings"})
// 	return
// }

// vector := models.Vector{
// 	ID:     fmt.Sprintf("text-%d", time.Now().UnixNano()), // Unique ID
// 	Values: vectorValues,
// }

// // Upsert vector to Pinecone with the provided namespace
// err = pkg.UpsertToPinecone([]models.Vector{vector}, request.Namespace)
// if err != nil {

// 	fmt.Println("error", err)
// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store vector in Pinecone"})
// 	return
// }
// 	var request struct {
// 		Query string `json:"query"`
// 		TopK  int    `json:"top_k"`
// 	}

// 	// Bind the JSON request body
// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
// 		return
// 	}

// 	// Get embedding for the query
// 	queryVector, err := pkg.GetEmbedding(request.Query, env.EmbeddingAPIKey)
// 	if err != nil {

// 		fmt.Println("error ", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate embeddings"})
// 		return
// 	}

// 	// Perform the search in Pinecone
// 	searchResults, err := pkg.SearchPinecone(queryVector, env.PineconeIndex, request.TopK)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"})
// 		return
// 	}

// 	// Respond with search results
// 	c.JSON(http.StatusOK, gin.H{"results": searchResults})
// })
