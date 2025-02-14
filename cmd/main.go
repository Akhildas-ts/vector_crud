package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"vectorchat/pkg"
	"vectorchat/pkg/config"
	"vectorchat/pkg/handler"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pinecone-io/go-pinecone/pinecone"
)

var (
	pc            *pinecone.Client
	idxConnection *pinecone.IndexConnection
	ctx           context.Context
)

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("config.Env ", cfg.ApiKey)

	pc, err = pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: cfg.ApiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create Pinecone client: %v", err)
	}

	ctx = context.Background()

	idxConnection, err = pc.Index(pinecone.NewIndexConnParams{
		Host:      cfg.PineconeHost,
		Namespace: "rag-namespace",
	})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection: %v", err)
	}
}

func main() {
	r := gin.Default()

	r.POST("/upsert", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		files := form.File["pdfs"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No PDF files provided"})
			return
		}

		// Process PDFs in parallel and store embeddings
		err = handler.Upsert(files, c)
		if err != nil {
			fmt.Println("Upsert error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upsert failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "PDFs processed successfully"})
	})
	r.POST("/search", pkg.SearchHandler(idxConnection, pc))
	r.POST("/query", handler.SearchHandler)

	log.Println("Starting server on port 8080...")
	r.Run(":8080")
	fmt.Println("working")
}
