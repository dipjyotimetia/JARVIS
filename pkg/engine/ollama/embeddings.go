package ollama

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
)

// GenerateEmbeddings generates embeddings using the modern Embed API
func (c *client) GenerateEmbeddings(ctx context.Context, model, prompt string) (*api.EmbedResponse, error) {
	if model == "" {
		model = getEmbeddingModel()
	}

	req := &api.EmbedRequest{
		Model:  model,
		Input:  prompt,
		Options: getEmbeddingOptions(),
	}

	resp, err := c.apiClient.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}

	return resp, nil
}

// GenerateEmbeddingsLegacy generates embeddings using the legacy Embeddings API
func (c *client) GenerateEmbeddingsLegacy(ctx context.Context, model string, input []string) (*api.EmbeddingResponse, error) {
	if model == "" {
		model = getEmbeddingModel()
	}

	req := &api.EmbeddingRequest{
		Model:   model,
		Prompt:  input[0], // Take first input for legacy API
		Options: getEmbeddingOptions(),
	}

	resp, err := c.apiClient.Embeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate legacy embeddings: %w", err)
	}

	return resp, nil
}

// GenerateEmbeddingsBatch generates embeddings for multiple inputs efficiently
func (c *client) GenerateEmbeddingsBatch(ctx context.Context, model string, inputs []string) ([]*api.EmbedResponse, error) {
	if model == "" {
		model = getEmbeddingModel()
	}

	var responses []*api.EmbedResponse
	
	for _, input := range inputs {
		resp, err := c.GenerateEmbeddings(ctx, model, input)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for input %q: %w", input, err)
		}
		responses = append(responses, resp)
	}

	return responses, nil
}

// CompareEmbeddings computes similarity between two text inputs using embeddings
func (c *client) CompareEmbeddings(ctx context.Context, model, text1, text2 string) (float64, error) {
	// Generate embeddings for both texts
	embed1, err := c.GenerateEmbeddings(ctx, model, text1)
	if err != nil {
		return 0, fmt.Errorf("failed to generate embedding for text1: %w", err)
	}

	embed2, err := c.GenerateEmbeddings(ctx, model, text2)
	if err != nil {
		return 0, fmt.Errorf("failed to generate embedding for text2: %w", err)
	}

	// Calculate cosine similarity (convert float32 to float64)
	vec1 := make([]float64, len(embed1.Embeddings[0]))
	vec2 := make([]float64, len(embed2.Embeddings[0]))
	
	for i, v := range embed1.Embeddings[0] {
		vec1[i] = float64(v)
	}
	for i, v := range embed2.Embeddings[0] {
		vec2[i] = float64(v)
	}
	
	similarity := cosineSimilarity(vec1, vec2)
	return similarity, nil
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// sqrt computes square root using Newton's method for simple dependency-free implementation
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}

	// Newton's method for square root
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}