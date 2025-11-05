package llm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// Global HTTP client with optimized settings for LLM API calls
var httpClient = &http.Client{
	Timeout: 60 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  false,
	},
}

// Provider represents the LLM provider type
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
	ProviderOllama    Provider = "ollama"
)

// Config holds LLM client configuration
type Config struct {
	Provider    Provider
	Model       string
	Temperature float64
	MaxTokens   int
	APIKey      string        // For cloud providers
	BaseURL     string        // For Ollama or custom endpoints
	TopP        float64
	TopK        int
	Timeout     time.Duration // Timeout for API calls (default: 60s)
	MaxRetries  int           // Maximum number of retries on failure (default: 3)
}

// Client provides a unified interface for multiple LLM providers
type Client struct {
	llm    llms.Model
	config Config
}

// New creates a new LLM client based on configuration
func New(ctx context.Context, config Config) (*Client, error) {
	// Auto-detect provider if not specified
	if config.Provider == "" {
		config.Provider = detectProvider()
	}

	// Set defaults
	if config.Model == "" {
		config.Model = getDefaultModel(config.Provider)
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 2048
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	var llmInstance llms.Model
	var err error

	switch config.Provider {
	case ProviderOpenAI:
		llmInstance, err = createOpenAI(config)
	case ProviderAnthropic:
		llmInstance, err = createAnthropic(config)
	case ProviderGoogle:
		llmInstance, err = createGoogle(config)
	case ProviderOllama:
		llmInstance, err = createOllama(config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", config.Provider, err)
	}

	return &Client{
		llm:    llmInstance,
		config: config,
	}, nil
}

// createOpenAI creates an OpenAI client
func createOpenAI(config Config) (llms.Model, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not provided")
	}

	opts := []openai.Option{
		openai.WithToken(apiKey),
		openai.WithModel(config.Model),
	}

	if config.BaseURL != "" {
		opts = append(opts, openai.WithBaseURL(config.BaseURL))
	}

	return openai.New(opts...)
}

// createAnthropic creates an Anthropic (Claude) client
func createAnthropic(config Config) (llms.Model, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not provided")
	}

	opts := []anthropic.Option{
		anthropic.WithToken(apiKey),
		anthropic.WithModel(config.Model),
	}

	return anthropic.New(opts...)
}

// createGoogle creates a Google AI (Gemini) client
func createGoogle(config Config) (llms.Model, error) {
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Google API key not provided")
	}

	opts := []googleai.Option{
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel(config.Model),
	}

	return googleai.New(context.Background(), opts...)
}

// createOllama creates an Ollama client
func createOllama(config Config) (llms.Model, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("OLLAMA_HOST")
	}
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	opts := []ollama.Option{
		ollama.WithModel(config.Model),
		ollama.WithServerURL(baseURL),
	}

	return ollama.New(opts...)
}

// detectProvider auto-detects the provider based on environment variables
func detectProvider() Provider {
	if os.Getenv("OPENAI_API_KEY") != "" {
		return ProviderOpenAI
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return ProviderAnthropic
	}
	if os.Getenv("GOOGLE_API_KEY") != "" {
		return ProviderGoogle
	}
	// Default to Ollama for local development
	return ProviderOllama
}

// getDefaultModel returns the default model for a provider
func getDefaultModel(provider Provider) string {
	defaults := map[Provider]string{
		ProviderOpenAI:    "gpt-4o-mini",
		ProviderAnthropic: "claude-3-5-sonnet-20241022",
		ProviderGoogle:    "gemini-1.5-flash",
		ProviderOllama:    "llama3.2",
	}

	if model := os.Getenv("LLM_MODEL"); model != "" {
		return model
	}

	return defaults[provider]
}

// Generate generates text from a prompt with timeout and retry logic
func (c *Client) Generate(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Apply default options
	defaultOpts := []llms.CallOption{
		llms.WithTemperature(c.config.Temperature),
		llms.WithMaxTokens(c.config.MaxTokens),
	}

	if c.config.TopP > 0 {
		defaultOpts = append(defaultOpts, llms.WithTopP(c.config.TopP))
	}
	if c.config.TopK > 0 {
		defaultOpts = append(defaultOpts, llms.WithTopK(c.config.TopK))
	}

	// Merge with provided options
	allOpts := append(defaultOpts, opts...)

	// Retry logic with exponential backoff
	var result string
	var err error
	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		result, err = llms.GenerateFromSinglePrompt(timeoutCtx, c.llm, prompt, allOpts...)
		if err == nil {
			return result, nil
		}

		// Check if context was cancelled or timed out
		if timeoutCtx.Err() != nil {
			return "", fmt.Errorf("generation failed after timeout: %w", err)
		}

		// Exponential backoff for retries
		if attempt < c.config.MaxRetries-1 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-time.After(backoff):
				continue
			case <-timeoutCtx.Done():
				return "", fmt.Errorf("generation failed: %w", timeoutCtx.Err())
			}
		}
	}

	return "", fmt.Errorf("generation failed after %d attempts: %w", c.config.MaxRetries, err)
}

// GenerateStream generates text with streaming support
func (c *Client) GenerateStream(ctx context.Context, prompt string, callback func(string) error, opts ...llms.CallOption) error {
	// Apply default options
	defaultOpts := []llms.CallOption{
		llms.WithTemperature(c.config.Temperature),
		llms.WithMaxTokens(c.config.MaxTokens),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			return callback(string(chunk))
		}),
	}

	if c.config.TopP > 0 {
		defaultOpts = append(defaultOpts, llms.WithTopP(c.config.TopP))
	}
	if c.config.TopK > 0 {
		defaultOpts = append(defaultOpts, llms.WithTopK(c.config.TopK))
	}

	// Merge with provided options
	allOpts := append(defaultOpts, opts...)

	_, err := llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, allOpts...)
	if err != nil {
		return fmt.Errorf("streaming generation failed: %w", err)
	}

	return nil
}

// Chat performs a chat completion
func (c *Client) Chat(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (string, error) {
	// Apply default options
	defaultOpts := []llms.CallOption{
		llms.WithTemperature(c.config.Temperature),
		llms.WithMaxTokens(c.config.MaxTokens),
	}

	if c.config.TopP > 0 {
		defaultOpts = append(defaultOpts, llms.WithTopP(c.config.TopP))
	}
	if c.config.TopK > 0 {
		defaultOpts = append(defaultOpts, llms.WithTopK(c.config.TopK))
	}

	// Merge with provided options
	allOpts := append(defaultOpts, opts...)

	result, err := c.llm.GenerateContent(ctx, messages, allOpts...)
	if err != nil {
		return "", fmt.Errorf("chat failed: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return result.Choices[0].Content, nil
}

// GetProvider returns the current provider
func (c *Client) GetProvider() Provider {
	return c.config.Provider
}

// GetModel returns the current model
func (c *Client) GetModel() string {
	return c.config.Model
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() Config {
	config := Config{
		Provider: Provider(strings.ToLower(os.Getenv("LLM_PROVIDER"))),
		Model:    os.Getenv("LLM_MODEL"),
		APIKey:   os.Getenv("LLM_API_KEY"),
		BaseURL:  os.Getenv("LLM_BASE_URL"),
	}

	// Parse temperature
	if temp := os.Getenv("LLM_TEMPERATURE"); temp != "" {
		fmt.Sscanf(temp, "%f", &config.Temperature)
	}

	// Parse max tokens
	if maxTokens := os.Getenv("LLM_MAX_TOKENS"); maxTokens != "" {
		fmt.Sscanf(maxTokens, "%d", &config.MaxTokens)
	}

	// Parse TopP
	if topP := os.Getenv("LLM_TOP_P"); topP != "" {
		fmt.Sscanf(topP, "%f", &config.TopP)
	}

	// Parse TopK
	if topK := os.Getenv("LLM_TOP_K"); topK != "" {
		fmt.Sscanf(topK, "%d", &config.TopK)
	}

	return config
}

// NewFromEnv creates a new LLM client from environment variables
func NewFromEnv(ctx context.Context) (*Client, error) {
	config := LoadConfigFromEnv()
	return New(ctx, config)
}
