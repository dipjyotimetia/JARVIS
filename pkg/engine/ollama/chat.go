package ollama

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
)

// Chat provides direct access to the Ollama Chat API with full control
func (c *client) Chat(ctx context.Context, req *api.ChatRequest, fn api.ChatResponseFunc) error {
	return c.apiClient.Chat(ctx, req, fn)
}

// ChatSimple provides a simple chat interface for single messages
func (c *client) ChatSimple(ctx context.Context, model, message string) (*api.ChatResponse, error) {
	if model == "" {
		model = getChatModel()
	}

	req := &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: message,
			},
		},
		Options: getChatOptions(),
	}

	var response *api.ChatResponse
	err := c.apiClient.Chat(ctx, req, func(resp api.ChatResponse) error {
		response = &resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("chat failed: %w", err)
	}

	return response, nil
}

// ChatWithHistory provides a chat interface that maintains conversation history
func (c *client) ChatWithHistory(ctx context.Context, model string, messages []api.Message) (*api.ChatResponse, error) {
	if model == "" {
		model = getChatModel()
	}

	req := &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Options:  getChatOptions(),
	}

	var response *api.ChatResponse
	err := c.apiClient.Chat(ctx, req, func(resp api.ChatResponse) error {
		response = &resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("chat with history failed: %w", err)
	}

	return response, nil
}

// ChatStream provides streaming chat for real-time responses
func (c *client) ChatStream(ctx context.Context, model, message string, fn api.ChatResponseFunc) error {
	if model == "" {
		model = getChatModel()
	}

	req := &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{
				Role:    "user",
				Content: message,
			},
		},
		Stream:  &[]bool{true}[0], // Enable streaming
		Options: getChatOptions(),
	}

	return c.apiClient.Chat(ctx, req, fn)
}

// ChatWithSystemPrompt creates a chat with a system prompt for behavior control
func (c *client) ChatWithSystemPrompt(ctx context.Context, model, systemPrompt, userMessage string) (*api.ChatResponse, error) {
	if model == "" {
		model = getChatModel()
	}

	messages := []api.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	return c.ChatWithHistory(ctx, model, messages)
}

// ConversationBuilder helps build complex conversations
type ConversationBuilder struct {
	model    string
	messages []api.Message
	options  map[string]any
}

// NewConversation creates a new conversation builder
func (c *client) NewConversation(model string) *ConversationBuilder {
	if model == "" {
		model = getChatModel()
	}

	return &ConversationBuilder{
		model:    model,
		messages: make([]api.Message, 0),
		options:  getChatOptions(),
	}
}

// SetSystemPrompt sets the system prompt for the conversation
func (cb *ConversationBuilder) SetSystemPrompt(prompt string) *ConversationBuilder {
	cb.messages = append([]api.Message{
		{
			Role:    "system",
			Content: prompt,
		},
	}, cb.messages...)
	return cb
}

// AddUserMessage adds a user message to the conversation
func (cb *ConversationBuilder) AddUserMessage(content string) *ConversationBuilder {
	cb.messages = append(cb.messages, api.Message{
		Role:    "user",
		Content: content,
	})
	return cb
}

// AddAssistantMessage adds an assistant message to the conversation
func (cb *ConversationBuilder) AddAssistantMessage(content string) *ConversationBuilder {
	cb.messages = append(cb.messages, api.Message{
		Role:    "assistant",
		Content: content,
	})
	return cb
}

// SetOptions sets custom options for the conversation
func (cb *ConversationBuilder) SetOptions(options map[string]any) *ConversationBuilder {
	cb.options = options
	return cb
}

// Execute runs the conversation and returns the response
func (cb *ConversationBuilder) Execute(ctx context.Context, client *client) (*api.ChatResponse, error) {
	req := &api.ChatRequest{
		Model:    cb.model,
		Messages: cb.messages,
		Options:  cb.options,
	}

	var response *api.ChatResponse
	err := client.apiClient.Chat(ctx, req, func(resp api.ChatResponse) error {
		response = &resp
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("conversation execution failed: %w", err)
	}

	return response, nil
}

// ExecuteStream runs the conversation with streaming responses
func (cb *ConversationBuilder) ExecuteStream(ctx context.Context, client *client, fn api.ChatResponseFunc) error {
	req := &api.ChatRequest{
		Model:    cb.model,
		Messages: cb.messages,
		Stream:   &[]bool{true}[0], // Enable streaming
		Options:  cb.options,
	}

	return client.apiClient.Chat(ctx, req, fn)
}