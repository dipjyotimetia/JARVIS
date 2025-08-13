package pact

import (
	"encoding/json"
	"strings"
	"testing"
)

// BenchmarkJSONParsing measures JSON parsing performance
func BenchmarkJSONParsing(b *testing.B) {
	sampleJSON := `{
		"consumer": {"name": "test-consumer"},
		"provider": {"name": "test-provider"},
		"interactions": [
			{
				"description": "test interaction",
				"request": {
					"method": "GET",
					"path": "/api/users",
					"headers": {"Accept": "application/json"}
				},
				"response": {
					"status": 200,
					"headers": {"Content-Type": "application/json"},
					"body": {"users": [{"id": 1, "name": "John"}]}
				}
			}
		]
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var contract map[string]interface{}
		json.Unmarshal([]byte(sampleJSON), &contract)
	}
}

// BenchmarkStringBuilderVsConcatenation compares string building methods
func BenchmarkStringBuilderVsConcatenation(b *testing.B) {
	parts := []string{
		"You are an expert in contract testing",
		"Generate a comprehensive Pact contract",
		"Requirements: Create realistic interactions",
		"Include proper HTTP methods, paths, headers",
		"Generate meaningful test data",
		"Use appropriate Pact matchers",
		"Follow Pact specification v3.0.0 format",
	}

	b.Run("StringBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			builder := stringBuilderPool.Get().(*strings.Builder)
			builder.Reset()
			for _, part := range parts {
				builder.WriteString(part)
				builder.WriteString("\\n")
			}
			result := builder.String()
			stringBuilderPool.Put(builder)
			_ = result
		}
	})

	b.Run("Concatenation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result string
			for _, part := range parts {
				result += part + "\\n"
			}
			_ = result
		}
	})

	b.Run("Join", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result := strings.Join(parts, "\\n")
			_ = result
		}
	})
}

// BenchmarkAIResponseParsing measures AI response parsing performance
func BenchmarkAIResponseParsing(b *testing.B) {
	g := &Generator{
		config: &GenerationConfig{
			ConsumerName: "test-consumer",
			ProviderName: "test-provider",
		},
	}

	aiResponse := `Here's the Pact contract:

	{
		"consumer": {"name": "test-consumer"},
		"provider": {"name": "test-provider"},
		"interactions": [
			{
				"description": "get user by id",
				"request": {
					"method": "GET",
					"path": "/users/123",
					"headers": {"Accept": "application/json"}
				},
				"response": {
					"status": 200,
					"headers": {"Content-Type": "application/json"},
					"body": {"id": 123, "name": "John Doe", "email": "john@example.com"}
				}
			}
		]
	}

	This contract defines the interaction between consumer and provider.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = g.parseAIResponse(aiResponse, nil)
	}
}

// BenchmarkPromptGeneration measures prompt generation performance
func BenchmarkPromptGeneration(b *testing.B) {
	config := &GenerationConfig{
		ConsumerName:    "test-consumer",
		ProviderName:    "test-provider",
		OutputPath:      "/tmp/test-pacts",
		SpecVersion:     "3.0.0",
		IncludeExamples: false,
	}

	specContent := []string{`openapi: "3.0.0"
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: Success`}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := &Generator{config: config}
		prompt := g.buildAIPrompt(nil, specContent)
		_ = prompt // Simulate AI call without actually calling
	}
}

// BenchmarkMapSerialization measures map serialization performance
func BenchmarkMapSerialization(b *testing.B) {
	contract := map[string]interface{}{
		"consumer": map[string]string{"name": "test-consumer"},
		"provider": map[string]string{"name": "test-provider"},
		"interactions": []map[string]interface{}{
			{
				"description": "get user by id",
				"request": map[string]interface{}{
					"method": "GET",
					"path":   "/users/123",
					"headers": map[string]string{
						"Accept": "application/json",
					},
				},
				"response": map[string]interface{}{
					"status": 200,
					"headers": map[string]string{
						"Content-Type": "application/json",
					},
					"body": map[string]interface{}{
						"id":    123,
						"name":  "John Doe",
						"email": "john@example.com",
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.MarshalIndent(contract, "", "  ")
	}
}

// BenchmarkJSONSanitization measures JSON sanitization performance
func BenchmarkJSONSanitization(b *testing.B) {
	g := &Generator{}
	dirtyJSON := "```json\\n{\\\"consumer\\\": {\\\"name\\\": \\\"test\\\"},\\\"provider\\\": {\\\"name\\\": \\\"api\\\"}}\\n```"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.sanitizeJSON(dirtyJSON)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}