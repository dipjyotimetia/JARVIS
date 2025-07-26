package ollama

import (
	"context"
	"testing"

	"github.com/ollama/ollama/api"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Ollama client: %v", err)
	}
	
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	
	// Test that client implements the interface
	var _ Client = client
	
	// Test Close method doesn't panic
	client.Close()
}

func TestNewWithURL(t *testing.T) {
	client, err := NewWithURL("http://localhost:11434")
	if err != nil {
		t.Fatalf("Failed to create Ollama client with URL: %v", err)
	}
	
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	
	// Test that client implements the interface
	var _ Client = client
	
	client.Close()
}

func TestDefaultModel(t *testing.T) {
	model := getDefaultModel()
	if model == "" {
		t.Fatal("Expected non-empty default model")
	}
	
	// Should return the default model name
	expected := "llama3.2"
	if model != expected {
		t.Errorf("Expected default model %s, got %s", expected, model)
	}
}

func TestChatModel(t *testing.T) {
	model := getChatModel()
	if model == "" {
		t.Fatal("Expected non-empty chat model")
	}
	
	// Should return the default model name since no chat-specific model is set
	expected := "llama3.2"
	if model != expected {
		t.Errorf("Expected chat model %s, got %s", expected, model)
	}
}

func TestVisionModel(t *testing.T) {
	model := getVisionModel()
	if model == "" {
		t.Fatal("Expected non-empty vision model")
	}
	
	// Should return the default vision model name
	expected := "llava"
	if model != expected {
		t.Errorf("Expected vision model %s, got %s", expected, model)
	}
}

func TestEmbeddingModel(t *testing.T) {
	model := getEmbeddingModel()
	if model == "" {
		t.Fatal("Expected non-empty embedding model")
	}
	
	// Should return the default embedding model name
	expected := "nomic-embed-text"
	if model != expected {
		t.Errorf("Expected embedding model %s, got %s", expected, model)
	}
}

func TestDefaultOptions(t *testing.T) {
	options := getDefaultOptions()
	if options == nil {
		t.Fatal("Expected non-nil options")
	}
	
	// Check that temperature is set
	if temp, ok := options["temperature"]; !ok || temp != 0.1 {
		t.Errorf("Expected temperature 0.1, got %v", temp)
	}
	
	// Check that top_k is set
	if topK, ok := options["top_k"]; !ok || topK != 40 {
		t.Errorf("Expected top_k 40, got %v", topK)
	}
	
	// Check that top_p is set
	if topP, ok := options["top_p"]; !ok || topP != 0.9 {
		t.Errorf("Expected top_p 0.9, got %v", topP)
	}
}

func TestChatOptions(t *testing.T) {
	options := getChatOptions()
	if options == nil {
		t.Fatal("Expected non-nil chat options")
	}
	
	// Check that temperature is higher for chat (more creative)
	if temp, ok := options["temperature"]; !ok || temp != 0.7 {
		t.Errorf("Expected chat temperature 0.7, got %v", temp)
	}
}

func TestEmbeddingOptions(t *testing.T) {
	options := getEmbeddingOptions()
	if options == nil {
		t.Fatal("Expected non-nil embedding options")
	}
	
	// Check that temperature is 0 for embeddings (deterministic)
	if temp, ok := options["temperature"]; !ok || temp != 0.0 {
		t.Errorf("Expected embedding temperature 0.0, got %v", temp)
	}
}

func TestGenerateRequest(t *testing.T) {
	req := &api.GenerateRequest{
		Model:  "test-model",
		Prompt: "test prompt",
		Stream: &[]bool{true}[0],
	}
	
	if req.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", req.Model)
	}
	
	if req.Prompt != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got %s", req.Prompt)
	}
	
	if req.Stream == nil || !*req.Stream {
		t.Error("Expected stream to be true")
	}
}

func TestChatRequest(t *testing.T) {
	req := &api.ChatRequest{
		Model: "test-model",
		Messages: []api.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
	}
	
	if req.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", req.Model)
	}
	
	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}
	
	if req.Messages[0].Role != "user" {
		t.Errorf("Expected role 'user', got %s", req.Messages[0].Role)
	}
	
	if req.Messages[0].Content != "Hello" {
		t.Errorf("Expected content 'Hello', got %s", req.Messages[0].Content)
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Test identical vectors
	vec1 := []float64{1.0, 2.0, 3.0}
	vec2 := []float64{1.0, 2.0, 3.0}
	
	similarity := cosineSimilarity(vec1, vec2)
	if similarity < 0.99 || similarity > 1.01 { // Allow for floating point precision
		t.Errorf("Expected similarity close to 1.0 for identical vectors, got %f", similarity)
	}
	
	// Test orthogonal vectors
	vec3 := []float64{1.0, 0.0}
	vec4 := []float64{0.0, 1.0}
	
	similarity2 := cosineSimilarity(vec3, vec4)
	if similarity2 < -0.01 || similarity2 > 0.01 { // Should be close to 0
		t.Errorf("Expected similarity close to 0.0 for orthogonal vectors, got %f", similarity2)
	}
}

func TestConversationBuilder(t *testing.T) {
	ctx := context.Background()
	c, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	clientImpl := c.(*client)
	conv := clientImpl.NewConversation("test-model")
	if conv == nil {
		t.Fatal("Expected non-nil conversation builder")
	}
	
	// Test building a conversation
	conv.SetSystemPrompt("You are a helpful assistant").
		AddUserMessage("Hello").
		AddAssistantMessage("Hi there!").
		AddUserMessage("How are you?")
	
	if len(conv.messages) != 4 {
		t.Errorf("Expected 4 messages, got %d", len(conv.messages))
	}
	
	if conv.messages[0].Role != "system" {
		t.Errorf("Expected first message to be system, got %s", conv.messages[0].Role)
	}
	
	if conv.messages[1].Role != "user" {
		t.Errorf("Expected second message to be user, got %s", conv.messages[1].Role)
	}
}