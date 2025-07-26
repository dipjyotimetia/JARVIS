package ollama

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/ollama/ollama/api"
)

// GenerateVision generates content using vision-capable models with images
func (c *client) GenerateVision(ctx context.Context, prompt string, imagePaths []string) (*api.GenerateResponse, error) {
	// Convert image paths to base64 encoded strings
	images, err := encodeImages(imagePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to encode images: %w", err)
	}

	req := &api.GenerateRequest{
		Model:   getVisionModel(),
		Prompt:  prompt,
		Images:  images,
		Stream:  &[]bool{false}[0], // Pointer to false for single response
		Options: getDefaultOptions(),
	}

	var finalResponse *api.GenerateResponse

	err = c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		finalResponse = &resp
		return nil
	})

	if err != nil {
		return nil, err
	}

	return finalResponse, nil
}

// GenerateVisionStream generates content with streaming response
func (c *client) GenerateVisionStream(ctx context.Context, prompt string) error {
	req := &api.GenerateRequest{
		Model:   getDefaultModel(),
		Prompt:  prompt,
		Stream:  &[]bool{true}[0], // Pointer to true
		Options: getDefaultOptions(),
	}

	return c.apiClient.Generate(ctx, req, func(resp api.GenerateResponse) error {
		fmt.Print(resp.Response)
		return nil
	})
}

// encodeImages converts image file paths to base64 encoded strings
func encodeImages(imagePaths []string) ([]api.ImageData, error) {
	var images []api.ImageData
	
	for _, path := range imagePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read image %s: %w", path, err)
		}
		
		encoded := base64.StdEncoding.EncodeToString(data)
		images = append(images, api.ImageData(encoded))
	}
	
	return images, nil
}