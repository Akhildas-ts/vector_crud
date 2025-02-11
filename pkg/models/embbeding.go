package models

type EmbeddingResponse [][]float32

type EmbeddingRequest struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}
