package testdata

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dipjyotimetia/jarvis/pkg/engine/ollama"
	"github.com/getkin/kin-openapi/openapi3"
)

// Generator handles AI-powered test data generation
type Generator struct {
	aiClient ollama.Client
}

// New creates a new test data generator
func New(ctx context.Context) (*Generator, error) {
	aiClient, err := ollama.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	return &Generator{
		aiClient: aiClient,
	}, nil
}

// GenerationOptions configures test data generation
type GenerationOptions struct {
	Count          int    // Number of test data samples to generate
	IncludeValid   bool   // Generate valid test cases
	IncludeInvalid bool   // Generate invalid/boundary test cases
	Format         string // Output format: json, yaml, csv
	Locale         string // Data locale (en-US, en-GB, etc.)
}

// TestDataResult represents generated test data
type TestDataResult struct {
	Endpoint    string                   `json:"endpoint"`
	Method      string                   `json:"method"`
	Description string                   `json:"description"`
	TestCases   []map[string]interface{} `json:"test_cases"`
}

// GenerateFromOpenAPI generates test data from OpenAPI specification
func (g *Generator) GenerateFromOpenAPI(ctx context.Context, specPath string, opts GenerationOptions) ([]TestDataResult, error) {
	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	var results []TestDataResult

	// Process each path and operation
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			result, err := g.generateForOperation(ctx, path, method, operation, doc, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to generate test data for %s %s: %w", method, path, err)
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// generateForOperation generates test data for a specific API operation
func (g *Generator) generateForOperation(
	ctx context.Context,
	path string,
	method string,
	operation *openapi3.Operation,
	doc *openapi3.T,
	opts GenerationOptions,
) (TestDataResult, error) {
	// Build prompt for AI
	prompt := g.buildPromptForOperation(path, method, operation, doc, opts)

	// Generate test data using AI
	response, err := g.aiClient.GenerateText(ctx, prompt)
	if err != nil {
		return TestDataResult{}, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse AI response into structured test data
	testCases, err := g.parseAIResponse(response.Response)
	if err != nil {
		return TestDataResult{}, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return TestDataResult{
		Endpoint:    path,
		Method:      method,
		Description: operation.Summary,
		TestCases:   testCases,
	}, nil
}

// buildPromptForOperation creates an AI prompt for test data generation
func (g *Generator) buildPromptForOperation(
	path string,
	method string,
	operation *openapi3.Operation,
	doc *openapi3.T,
	opts GenerationOptions,
) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Generate %d realistic test data samples for the following API endpoint:\n\n", opts.Count))
	builder.WriteString(fmt.Sprintf("Endpoint: %s %s\n", strings.ToUpper(method), path))

	if operation.Summary != "" {
		builder.WriteString(fmt.Sprintf("Summary: %s\n", operation.Summary))
	}

	if operation.Description != "" {
		builder.WriteString(fmt.Sprintf("Description: %s\n", operation.Description))
	}

	// Add request body schema if present
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		if content, ok := operation.RequestBody.Value.Content["application/json"]; ok {
			if content.Schema != nil && content.Schema.Value != nil {
				schemaJSON, _ := json.MarshalIndent(content.Schema.Value, "", "  ")
				builder.WriteString(fmt.Sprintf("\nRequest Body Schema:\n%s\n", string(schemaJSON)))
			}
		}
	}

	// Add parameters
	if len(operation.Parameters) > 0 {
		builder.WriteString("\nParameters:\n")
		for _, param := range operation.Parameters {
			if param.Value != nil {
				builder.WriteString(fmt.Sprintf("- %s (%s, %s): %s\n",
					param.Value.Name,
					param.Value.In,
					func() string {
						if param.Value.Required {
							return "required"
						}
						return "optional"
					}(),
					param.Value.Description,
				))
			}
		}
	}

	// Add generation requirements
	builder.WriteString("\nRequirements:\n")
	if opts.IncludeValid {
		builder.WriteString("- Include valid test cases that should succeed\n")
	}
	if opts.IncludeInvalid {
		builder.WriteString("- Include boundary values, edge cases, and invalid inputs\n")
		builder.WriteString("- Include tests for validation errors\n")
	}
	if opts.Locale != "" {
		builder.WriteString(fmt.Sprintf("- Use %s locale for data (dates, phone numbers, addresses)\n", opts.Locale))
	}

	builder.WriteString("\nOutput Format:\n")
	builder.WriteString("Return ONLY a JSON array of test case objects. Each object should contain:\n")
	builder.WriteString("- \"name\": A descriptive name for the test case\n")
	builder.WriteString("- \"type\": Either \"valid\" or \"invalid\"\n")
	builder.WriteString("- \"data\": The actual test data matching the schema\n")
	builder.WriteString("- \"expected_status\": Expected HTTP status code\n")
	builder.WriteString("- \"description\": Brief explanation of what this test validates\n")
	builder.WriteString("\nReturn ONLY the JSON array, no additional text or explanation.")

	return builder.String()
}

// parseAIResponse parses the AI-generated response into test cases
func (g *Generator) parseAIResponse(response string) ([]map[string]interface{}, error) {
	// Clean up the response - remove markdown code blocks if present
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var testCases []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &testCases); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return testCases, nil
}

// GenerateForSchema generates test data for a specific JSON schema
func (g *Generator) GenerateForSchema(ctx context.Context, schemaJSON string, opts GenerationOptions) ([]map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Generate %d realistic test data samples matching this JSON schema:

%s

Requirements:
`, opts.Count, schemaJSON)

	if opts.IncludeValid {
		prompt += "- Include valid test cases\n"
	}
	if opts.IncludeInvalid {
		prompt += "- Include boundary values and invalid inputs\n"
	}

	prompt += `
Return ONLY a JSON array of test data objects matching the schema.
No additional text or explanation.`

	response, err := g.aiClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return g.parseAIResponse(response.Response)
}

// GenerateFromExample generates variations of test data from an example
func (g *Generator) GenerateFromExample(ctx context.Context, exampleJSON string, count int) ([]map[string]interface{}, error) {
	prompt := fmt.Sprintf(`Given this example data:

%s

Generate %d similar but different variations of this data.
Keep the same structure but vary the values realistically.
Return ONLY a JSON array of objects, no additional text.`, exampleJSON, count)

	response, err := g.aiClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	return g.parseAIResponse(response.Response)
}
