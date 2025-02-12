package handler

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"sync"
	"time"
	"vectorchat/pkg"
	"vectorchat/pkg/config"
	"vectorchat/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

func Upsert(files []*multipart.FileHeader, c *gin.Context) error {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: cfg.PineconeApiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create Pinecone client: %w", err)
	}

	indexName := "dimensions1024"
	idxModel, err := client.DescribeIndex(ctx, indexName)
	if err != nil {
		return fmt.Errorf("failed to describe index \"%v\": %w", indexName, err)
	}

	idxConnection, err := client.Index(pinecone.NewIndexConnParams{
		Host:      idxModel.Host,
		Namespace: "example-namespace",
	})
	if err != nil {
		return fmt.Errorf("failed to create IndexConnection: %w", err)
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		records []*pinecone.Vector
	)

	for _, file := range files {
		wg.Add(1)
		go func(file *multipart.FileHeader) {
			defer wg.Done()
			vector, err := processFile(file, c, client)
			if err != nil {
				fmt.Println("Error processing file:", err)
				return
			}
			mu.Lock()
			records = append(records, vector)
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	batchSize := 10
	for i := 0; i < len(records); i += batchSize {
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		start := time.Now()
		count, err := idxConnection.UpsertVectors(ctx, records[i:end])
		fmt.Println("Upsert took:", time.Since(start))

		if err != nil {
			return fmt.Errorf("failed to upsert vectors: %w", err)
		}
		fmt.Printf("Successfully upserted %d vector(s)\n", count)
	}

	return nil
}

func processFile(file *multipart.FileHeader, c *gin.Context, client *pinecone.Client) (*pinecone.Vector, error) {
	// Save file temporarily
	tempFile, err := os.CreateTemp("", "upload-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if err := c.SaveUploadedFile(file, tempFile.Name()); err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}

	text, err := extractTextFromPDF(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to extract text from PDF: %w", err)
	}

	chunks := helper.SplitText(text, 7000)
	var summaries []string

	for _, chunk := range chunks {
		start := time.Now()
		summary, err := pkg.SearchOpenAI(chunk)
		fmt.Println("OpenAI summary took:", time.Since(start))
		if err != nil {
			return nil, fmt.Errorf("failed to get summary from OpenAI: %w", err)
		}
		summaries = append(summaries, summary)
	}

	summaryText := strings.Join(summaries, " ")

	// Generate embeddings using Pinecone
	ctx := context.Background()
	start := time.Now()
	docEmbeddingsResponse, err := client.Inference.Embed(ctx, &pinecone.EmbedRequest{
		Model:      "multilingual-e5-large",
		TextInputs: []string{text},
		Parameters: pinecone.EmbedParameters{InputType: "passage", Truncate: "END"},
	})
	fmt.Println("Embedding took:", time.Since(start))

	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	embedding := (*docEmbeddingsResponse.Data)[0]
	values := embedding.Values

	metadataMap := map[string]interface{}{
		"filename": file.Filename,
		"text":     summaryText,
	}
	metadata, _ := structpb.NewStruct(metadataMap)

	return &pinecone.Vector{
		Id:       file.Filename,
		Values:   *values,
		Metadata: metadata,
	}, nil
}

func extractTextFromPDF(pdfPath string) (string, error) {
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	textChan := make(chan string, totalPages)
	errChan := make(chan error, totalPages)
	var wg sync.WaitGroup

	// Process each page concurrently
	for i := 1; i <= totalPages; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			pageText, err := r.Page(page).GetPlainText(nil)
			if err != nil {
				errChan <- fmt.Errorf("failed to extract text from page %d: %w", page, err)
				return
			}
			textChan <- pageText
		}(i)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(textChan)
		close(errChan)
	}()

	var text strings.Builder
	for t := range textChan {
		text.WriteString(t)
		text.WriteString("\n")
	}

	// Check if there were errors
	if err := <-errChan; err != nil {
		return "", err
	}

	return text.String(), nil
}
