package pkg

import (
	"context"
	"vectorchat/pkg/config"
	"vectorchat/pkg/models"

	"fmt"
	"log"

	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

//just for make struct prety

// func prettifyStruct(obj interface{}) string {
// 	bytes, _ := json.MarshalIndent(obj, "", "  ")
// 	return string(bytes)
// // }

// func extractTextFromPDF(pdfPath string) (string, error) {
// 	f, r, err := pdf.Open(pdfPath)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer f.Close()

// 	var buf bytes.Buffer
// 	b, err := r.GetPlainText()
// 	if err != nil {
// 		return "", err
// 	}

// 	buf.ReadFrom(b)
// 	return buf.String(), nil
// }

func Upsert(client *pinecone.Client, req []models.EmbeddingRequest, ctx context.Context) error {

	// Specify the embedding model and parameters
	embeddingModel := "multilingual-e5-large"

	docParameters := pinecone.EmbedParameters{
		InputType: "passage",
		Truncate:  "END",
	}

	var documents []string
	for _, d := range req {
		documents = append(documents, d.Data)
	}

	docEmbeddingsResponse, err := client.Inference.Embed(ctx, &pinecone.EmbedRequest{
		Model:      embeddingModel,
		TextInputs: documents,
		Parameters: docParameters,
	})
	if err != nil {

		log.Fatalf("Failed to embed documents: %v", err)
		return fmt.Errorf("error is ", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {

		return fmt.Errorf("load config error", err)
	}
	indexName := cfg.PineconeIndex

	idxModel, err := client.DescribeIndex(ctx, indexName)
	if err != nil {
		log.Fatalf("Failed to describe index \"%v\": %v", indexName, err)
	}

	idxConnection, err := client.Index(pinecone.NewIndexConnParams{Host: idxModel.Host, Namespace: "example-namespace"})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection for Host %v: %v", idxModel.Host, err)
	}

	var records []*pinecone.Vector
	for i := range req {
		metadataMap := map[string]interface{}{
			"text": req[i].Data,
		}
		metadata, err := structpb.NewStruct(metadataMap)
		if err != nil {
			log.Fatalf("Failed to create metadata map. Error: %v", err)
		}

		embedding := (*docEmbeddingsResponse.Data)[i]
		values := embedding.Values
		records = append(records, &pinecone.Vector{
			Id:       req[i].Id,
			Values:   *values,
			Metadata: metadata,
		})
	}
	count, err := idxConnection.UpsertVectors(ctx, records)
	if err != nil {
		log.Fatalf("Failed to upsert vectors: %v", err)

		return fmt.Errorf("failed to upsert ", err)
	} else {
		fmt.Printf("Successfully upserted %d vector(s)!\n", count)
	}

	return nil

}


// func UpsertToPinecone(vectors []models.Vector, namespace string) error {
// 	payload := models.PineconeUpsertRequest{
// 		Vectors:   vectors,
// 		Namespace: namespace,
// 	}

// 	jsonData, err := json.Marshal(payload)
// 	if err != nil {
// 		log.Printf("Error marshalling upsert request: %v", err)
// 		return err
// 	}

// 	req, err := http.NewRequest("POST", fmt.Sprintf("%s/vectors/upsert", env.PineconeHost), bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		log.Printf("Error creating HTTP request: %v", err)
// 		return err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Api-Key", env.PineconeAPIKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Printf("Error executing API request: %v", err)
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := ioutil.ReadAll(resp.Body)
// 		log.Printf("Pinecone upsert failed with status %d: %s", resp.StatusCode, string(body))
// 		return fmt.Errorf("failed to upsert data, status code: %d", resp.StatusCode)
// 	}

// 	return nil
// }
