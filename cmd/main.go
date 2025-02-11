package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"vectorchat/pkg"
	"vectorchat/pkg/config"
	"vectorchat/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ledongthuc/pdf"
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
		Namespace: "example-namespace",
	})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection: %v", err)
	}
}

func main() {
	r := gin.Default()

	r.POST("/upsert", func(c *gin.Context) {
		var request []models.EmbeddingRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if err := pkg.Upsert(pc, request, ctx); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Upsert issue"})
			return
		}
	})

	r.POST("/search", pkg.SearchHandler(idxConnection, pc))
	r.POST("/convert", ConvertMultiplePDFsToTextArray)
	r.POST("/query", SearchHandler)

	log.Println("Starting server on port 8080...")
	r.Run(":8080")
}

func SearchHandler(c *gin.Context) {

	var data interface{}
	// var request map[string]string
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var v map[string]interface{}
	switch temp := data.(type) {
	case map[string]interface{}:
		v = temp // Assign temp to v
		fmt.Println(v)
	}

	fmt.Println("data", data)

	result, err := pkg.SearchGemini(v["query"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Println("a2")

	c.JSON(http.StatusOK, gin.H{"result": result})
}

func ConvertMultiplePDFsToTextArray(c *gin.Context) {
	// Step 1: Receive multiple files from the request
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get form data"})
		return
	}

	files := form.File["pdfs[]"] // Expecting multiple files under key "pdfs[]"
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No PDF files uploaded"})
		return
	}

	// Step 2: Iterate over each PDF file
	var results []gin.H

	for _, file := range files {
		tempFile := fmt.Sprintf("/tmp/%s", file.Filename)
		if err := c.SaveUploadedFile(file, tempFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PDF"})
			return
		}

		// Open the saved PDF file
		f, err := os.Open(tempFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to open %s", file.Filename)})
			return
		}
		defer f.Close()

		// Read the PDF
		reader, err := pdf.NewReader(f, file.Size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to read %s", file.Filename)})
			return
		}

		// Extract text from each page
		var pagesText []string
		numPages := reader.NumPage()
		for i := 1; i <= numPages; i++ {
			page := reader.Page(i)
			if page.V.IsNull() {
				continue
			}

			text, err := page.GetPlainText(nil)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to extract text from %s", file.Filename)})
				return
			}

			pagesText = append(pagesText, text)
		}

		// Store results for each PDF
		results = append(results, gin.H{
			"file":  file.Filename,
			"pages": pagesText,
		})
	}

	// Step 3: Return the extracted text for all PDFs as JSON
	c.JSON(http.StatusOK, gin.H{"results": results})
}
