
package models

type PineconeUpsertRequest struct {
	Vectors   []Vector `json:"vectors"`
	Namespace string   `json:"namespace"`
}

type Vector struct {
	ID     string    `json:"id"`
	Values []float32 `json:"values"`
}

type PineconeSearchRequest struct {
	Vector    []float32 `json:"vector"`   // Vector to search with
	TopK      int       `json:"top_k" `    // Number of nearest neighbors to return
	Namespace string    `json:"namespace"` // The namespace to search in
}

// PineconeSearchResponse represents the structure of the search response.
type PineconeSearchResponse struct {
	ID     string    `json:"id"`
	Score  float32   `json:"score"`
	Values []float32 `json:"values"`
}

// PineconeSearchResult is the wrapper for the search result.
type PineconeSearchResult struct {
	Results []PineconeSearchResponse `json:"results"`
}

type PineconeSearchMatch struct {
	ID     string    `json:"id"`
	Score  float32   `json:"score"`
	Values []float32 `json:"values"`
}

type PineconeEmbeddingRequest struct {
	Inputs []string `json:"inputs"`
	Model  string   `json:"model"`
}

// Fix: Adjusted the struct to match the API response format.
type PineconeEmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}
