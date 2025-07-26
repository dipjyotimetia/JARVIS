package ollama

import (
	"context"
	"os"

	"github.com/ollama/ollama/api"
)

const (
	DefaultHost = "http://localhost:11434"
)

// Client provides a comprehensive interface for interacting with Ollama
type Client interface {
	// Generation APIs
	GenerateText(ctx context.Context, prompt string) (*api.GenerateResponse, error)
	GenerateTextStream(ctx context.Context, specs []string, specType string) error
	GenerateTextStreamWriter(ctx context.Context, specs []string, language, specType string, outputFolder string) error
	GenerateVision(ctx context.Context, prompt string, images []string) (*api.GenerateResponse, error)
	GenerateVisionStream(ctx context.Context, prompt string) error
	GenerateWithOptions(ctx context.Context, req *api.GenerateRequest, fn api.GenerateResponseFunc) error

	// Chat APIs - Modern conversational interface
	Chat(ctx context.Context, req *api.ChatRequest, fn api.ChatResponseFunc) error
	ChatSimple(ctx context.Context, model, message string) (*api.ChatResponse, error)
	ChatWithHistory(ctx context.Context, model string, messages []api.Message) (*api.ChatResponse, error)

	// Embeddings APIs
	GenerateEmbeddings(ctx context.Context, model, prompt string) (*api.EmbedResponse, error)
	GenerateEmbeddingsLegacy(ctx context.Context, model string, input []string) (*api.EmbeddingResponse, error)

	// Model Management APIs
	ListModels(ctx context.Context) (*api.ListResponse, error)
	ListRunningModels(ctx context.Context) (*api.ProcessResponse, error)
	ShowModelInfo(ctx context.Context, modelName string) (*api.ShowResponse, error)
	IsModelAvailable(ctx context.Context, modelName string) (bool, error)
	PullModel(ctx context.Context, modelName string, fn api.PullProgressFunc) error
	PushModel(ctx context.Context, modelName string, fn api.PushProgressFunc) error
	CopyModel(ctx context.Context, source, destination string) error
	DeleteModel(ctx context.Context, modelName string) error

	// System APIs
	Heartbeat(ctx context.Context) error
	Version(ctx context.Context) (string, error)

	// Legacy methods for backward compatibility
	Close()
}

type client struct {
	apiClient *api.Client
}

// New creates a new Ollama client using the official API library
func New(ctx context.Context) (Client, error) {
	// Create client from environment (OLLAMA_HOST)
	apiClient, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &client{
		apiClient: apiClient,
	}, nil
}

// NewWithURL creates a new Ollama client with a specific URL
func NewWithURL(baseURL string) (Client, error) {
	apiClient := api.NewClient(nil, nil)

	return &client{
		apiClient: apiClient,
	}, nil
}

// Close closes the Ollama client
func (c *client) Close() {
	// The official API client doesn't require explicit closing
}

// getDefaultModel returns a default model name
func getDefaultModel() string {
	if model := os.Getenv("OLLAMA_MODEL"); model != "" {
		return model
	}
	return "llama3.2" // Default to a common model
}

// getChatModel returns a model optimized for chat
func getChatModel() string {
	if model := os.Getenv("OLLAMA_CHAT_MODEL"); model != "" {
		return model
	}
	return getDefaultModel() // Fallback to default model
}

// getVisionModel returns a vision-capable model name
func getVisionModel() string {
	if model := os.Getenv("OLLAMA_VISION_MODEL"); model != "" {
		return model
	}
	return "llava" // Default vision model
}

// getEmbeddingModel returns a model optimized for embeddings
func getEmbeddingModel() string {
	if model := os.Getenv("OLLAMA_EMBEDDING_MODEL"); model != "" {
		return model
	}
	return "nomic-embed-text" // Default embedding model
}

// getDefaultOptions returns default generation options
func getDefaultOptions() map[string]any {
	return map[string]any{
		"temperature": 0.1,
		"top_k":       40,
		"top_p":       0.9,
	}
}

// getChatOptions returns optimized options for chat
func getChatOptions() map[string]any {
	return map[string]any{
		"temperature": 0.7, // Higher for more creative responses
		"top_k":       40,
		"top_p":       0.9,
	}
}

// getEmbeddingOptions returns optimized options for embeddings
func getEmbeddingOptions() map[string]any {
	return map[string]any{
		"temperature": 0.0, // Deterministic for embeddings
	}
}