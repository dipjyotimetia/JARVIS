package ollama

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

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

	// Configuration methods
	UpdateConfig(taskType string, config AIConfig) error
	GetConfig(taskType string) AIConfig
	
	// Legacy methods for backward compatibility
	Close()
}

type client struct {
	apiClient *api.Client
	configs   map[string]AIConfig
	configMu  sync.RWMutex
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
		configs:   make(map[string]AIConfig),
	}, nil
}

// NewWithURL creates a new Ollama client with a specific URL
func NewWithURL(baseURL string) (Client, error) {
	apiClient := api.NewClient(nil, nil)

	return &client{
		apiClient: apiClient,
		configs:   make(map[string]AIConfig),
	}, nil
}

// UpdateConfig updates the configuration for a specific task type
func (c *client) UpdateConfig(taskType string, config AIConfig) error {
	c.configMu.Lock()
	defer c.configMu.Unlock()
	
	c.configs[taskType] = config
	return nil
}

// GetConfig returns the configuration for a specific task type
func (c *client) GetConfig(taskType string) AIConfig {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	
	if config, exists := c.configs[taskType]; exists {
		return config
	}
	
	// Return default configuration if not found
	return loadAIConfig(taskType)
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

// AIConfig holds configuration for AI operations
type AIConfig struct {
	Temperature    float64                `json:"temperature" yaml:"temperature"`
	TopK          int                    `json:"top_k" yaml:"top_k"`
	TopP          float64                `json:"top_p" yaml:"top_p"`
	NumCtx        int                    `json:"num_ctx" yaml:"num_ctx"`
	TimeoutSeconds int                    `json:"timeout_seconds" yaml:"timeout_seconds"`
	CustomOptions map[string]interface{} `json:"custom_options" yaml:"custom_options"`
}

// getDefaultOptions returns default generation options with performance tuning
func getDefaultOptions() map[string]any {
	return getConfigurableOptions("generation")
}

// getChatOptions returns optimized options for chat
func getChatOptions() map[string]any {
	return getConfigurableOptions("chat")
}

// getEmbeddingOptions returns optimized options for embeddings
func getEmbeddingOptions() map[string]any {
	return getConfigurableOptions("embedding")
}

// getConfigurableOptions provides full control over generation parameters
func getConfigurableOptions(taskType string) map[string]any {
	// Load configuration for task type
	config := loadAIConfig(taskType)
	
	options := map[string]any{
		"temperature": config.Temperature,
		"top_k":       config.TopK,
		"top_p":       config.TopP,
		"num_ctx":     config.NumCtx,
		"timeout":     config.TimeoutSeconds,
	}
	
	// Merge custom options
	for k, v := range config.CustomOptions {
		options[k] = v
	}
	
	return options
}

// getConfigurableOptionsForClient provides options for a specific client instance
func (c *client) getConfigurableOptionsForClient(taskType string) map[string]any {
	c.configMu.RLock()
	defer c.configMu.RUnlock()
	
	var config AIConfig
	if clientConfig, exists := c.configs[taskType]; exists {
		config = clientConfig
	} else {
		config = loadAIConfig(taskType)
	}
	
	options := map[string]any{
		"temperature": config.Temperature,
		"top_k":       config.TopK,
		"top_p":       config.TopP,
		"num_ctx":     config.NumCtx,
		"timeout":     config.TimeoutSeconds,
	}
	
	// Merge custom options
	for k, v := range config.CustomOptions {
		options[k] = v
	}
	
	return options
}

// loadAIConfig loads AI configuration with environment variable overrides
func loadAIConfig(taskType string) AIConfig {
	// Default configurations for different task types
	defaults := map[string]AIConfig{
		"generation": {
			Temperature:    0.1,
			TopK:          40,
			TopP:          0.9,
			NumCtx:        4096,
			TimeoutSeconds: 300,
			CustomOptions:  make(map[string]interface{}),
		},
		"chat": {
			Temperature:    0.4,
			TopK:          50,
			TopP:          0.9,
			NumCtx:        8192,
			TimeoutSeconds: 600,
			CustomOptions:  make(map[string]interface{}),
		},
		"embedding": {
			Temperature:    0.0,
			TopK:          1,
			TopP:          1.0,
			NumCtx:        2048,
			TimeoutSeconds: 30,
			CustomOptions:  make(map[string]interface{}),
		},
	}
	
	config, exists := defaults[taskType]
	if !exists {
		config = defaults["generation"] // fallback
	}
	
	// Override from environment variables
	prefix := "OLLAMA_" + strings.ToUpper(taskType) + "_"
	
	if temp := os.Getenv(prefix + "TEMPERATURE"); temp != "" {
		if f, err := strconv.ParseFloat(temp, 64); err == nil {
			config.Temperature = f
		}
	}
	
	if topK := os.Getenv(prefix + "TOP_K"); topK != "" {
		if i, err := strconv.Atoi(topK); err == nil {
			config.TopK = i
		}
	}
	
	if topP := os.Getenv(prefix + "TOP_P"); topP != "" {
		if f, err := strconv.ParseFloat(topP, 64); err == nil {
			config.TopP = f
		}
	}
	
	if numCtx := os.Getenv(prefix + "NUM_CTX"); numCtx != "" {
		if i, err := strconv.Atoi(numCtx); err == nil {
			config.NumCtx = i
		}
	}
	
	if timeout := os.Getenv(prefix + "TIMEOUT"); timeout != "" {
		if i, err := strconv.Atoi(timeout); err == nil {
			config.TimeoutSeconds = i
		}
	}
	
	// Load global overrides
	if temp := os.Getenv("OLLAMA_TEMPERATURE"); temp != "" {
		if f, err := strconv.ParseFloat(temp, 64); err == nil {
			config.Temperature = f
		}
	}
	
	return config
}