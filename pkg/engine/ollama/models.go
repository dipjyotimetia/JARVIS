package ollama

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
)

// ListModels returns the list of available models
func (c *client) ListModels(ctx context.Context) (*api.ListResponse, error) {
	return c.apiClient.List(ctx)
}

// ShowModelInfo returns detailed information about a specific model
func (c *client) ShowModelInfo(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	return c.apiClient.Show(ctx, &api.ShowRequest{
		Name: modelName,
	})
}

// ListRunningModels returns the list of currently running models
func (c *client) ListRunningModels(ctx context.Context) (*api.ProcessResponse, error) {
	return c.apiClient.ListRunning(ctx)
}

// IsModelAvailable checks if a model is available locally
func (c *client) IsModelAvailable(ctx context.Context, modelName string) (bool, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false, err
	}

	for _, model := range models.Models {
		if model.Name == modelName {
			return true, nil
		}
	}

	return false, nil
}

// PullModel downloads a model from a registry
func (c *client) PullModel(ctx context.Context, modelName string, fn api.PullProgressFunc) error {
	req := &api.PullRequest{
		Name: modelName,
	}
	return c.apiClient.Pull(ctx, req, fn)
}

// PushModel uploads a model to a registry
func (c *client) PushModel(ctx context.Context, modelName string, fn api.PushProgressFunc) error {
	req := &api.PushRequest{
		Name: modelName,
	}
	return c.apiClient.Push(ctx, req, fn)
}

// CopyModel creates a copy of a model with a new name
func (c *client) CopyModel(ctx context.Context, source, destination string) error {
	req := &api.CopyRequest{
		Source:      source,
		Destination: destination,
	}
	return c.apiClient.Copy(ctx, req)
}

// DeleteModel removes a model from local storage
func (c *client) DeleteModel(ctx context.Context, modelName string) error {
	req := &api.DeleteRequest{
		Name: modelName,
	}
	return c.apiClient.Delete(ctx, req)
}

// PullModelSimple downloads a model with basic progress reporting
func (c *client) PullModelSimple(ctx context.Context, modelName string) error {
	return c.PullModel(ctx, modelName, func(resp api.ProgressResponse) error {
		if resp.Status != "" {
			fmt.Printf("Pulling %s: %s\n", modelName, resp.Status)
		}
		return nil
	})
}

// Common model names
const (
	ModelLlama32    = "llama3.2"
	ModelLlama31    = "llama3.1"
	ModelLlama3     = "llama3"
	ModelCodellama  = "codellama"
	ModelLlava      = "llava"          // Vision model
	ModelBakllava   = "bakllava"       // Another vision model
	ModelGemma      = "gemma"
	ModelMistral    = "mistral"
	ModelNeuralChat = "neural-chat"
	ModelStarCoder  = "starcoder"
)