package pact

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/ollama"
	"github.com/getkin/kin-openapi/openapi3"
)

// Buffer pools for optimization
var (
	stringBuilderPool = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
	
	jsonBufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 4096)
		},
	}
)

// Generator handles Pact contract generation
type Generator struct {
	config *GenerationConfig
	ai     ollama.Client
}

// NewGenerator creates a new Pact generator
func NewGenerator(ctx context.Context, config *GenerationConfig) (*Generator, error) {
	ai, err := ollama.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI client: %w", err)
	}

	if config.SpecVersion == "" {
		config.SpecVersion = "3.0.0"
	}

	return &Generator{
		config: config,
		ai:     ai,
	}, nil
}

// GenerateFromOpenAPI generates Pact contracts from OpenAPI specification
func (g *Generator) GenerateFromOpenAPI(ctx context.Context, specPath string) (*ContractGenerationResult, error) {
	// Read the OpenAPI specification
	specFiles, err := files.ListFiles(specPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list spec files: %w", err)
	}

	if len(specFiles) == 0 {
		return nil, fmt.Errorf("no specification files found at path: %s", specPath)
	}

	specContent, err := files.ReadFile(specFiles[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read spec file: %w", err)
	}

	// Parse OpenAPI spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(strings.Join(specContent, "\n")))
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Generate contract using AI
	contract, err := g.generateContractFromSpec(ctx, doc, specContent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate contract: %w", err)
	}

	// Save contract to file
	filePath, err := g.saveContract(contract)
	if err != nil {
		return nil, fmt.Errorf("failed to save contract: %w", err)
	}

	// Generate test code if requested
	var testCode string
	if g.config.IncludeExamples {
		testCode, err = g.generateTestCode(ctx, contract)
		if err != nil {
			// Log warning but don't fail
			slog.Warn("Failed to generate test code", "error", err)
		}
	}

	return &ContractGenerationResult{
		Contract:     contract,
		FilePath:     filePath,
		TestCode:     testCode,
		Language:     g.config.Language,
		Framework:    g.config.Framework,
		GeneratedAt:  time.Now(),
		SourceSpec:   specFiles[0],
		Interactions: len(contract.Interactions),
	}, nil
}

// generateContractFromSpec uses AI to generate Pact contract from OpenAPI spec
func (g *Generator) generateContractFromSpec(ctx context.Context, doc *openapi3.T, specContent []string) (*PactContract, error) {
	prompt := g.buildAIPrompt(doc, specContent)
	
	// Generate interactions using AI
	aiResponse, err := g.ai.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}
	response := aiResponse.Response

	// Parse AI response and create contract
	contract, err := g.parseAIResponse(response, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Set contract metadata
	contract.Consumer.Name = g.config.ConsumerName
	contract.Provider.Name = g.config.ProviderName
	contract.SetMetadata(g.config.SpecVersion, "jarvis-pact-generator", "1.0.0")

	return contract, nil
}

// buildAIPrompt creates a prompt for the AI to generate Pact contracts using pooled string builder
func (g *Generator) buildAIPrompt(_ *openapi3.T, specContent []string) string {
	builder := stringBuilderPool.Get().(*strings.Builder)
	defer stringBuilderPool.Put(builder)
	builder.Reset()
	
	builder.WriteString("You are an expert in contract testing and Pact specification. ")
	builder.WriteString("Generate a comprehensive Pact contract from the following OpenAPI specification.\n\n")
	
	builder.WriteString("Requirements:\n")
	builder.WriteString("- Create realistic interactions for each endpoint\n")
	builder.WriteString("- Include proper HTTP methods, paths, headers, and response codes\n")
	builder.WriteString("- Generate meaningful test data for request/response bodies\n")
	builder.WriteString("- Include both success and error scenarios\n")
	builder.WriteString("- Use appropriate Pact matchers for flexible matching\n")
	builder.WriteString("- Follow Pact specification v3.0.0 format\n\n")
	
	if g.config.Language != "" {
		builder.WriteString("Target language: ")
		builder.WriteString(g.config.Language)
		builder.WriteString("\n")
	}
	if g.config.Framework != "" {
		builder.WriteString("Target framework: ")
		builder.WriteString(g.config.Framework)
		builder.WriteString("\n")
	}
	
	for key, value := range g.config.ExtraContext {
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(value)
		builder.WriteString("\n")
	}
	
	builder.WriteString("\nOpenAPI Specification:\n")
	for _, content := range specContent {
		builder.WriteString(content)
		builder.WriteString("\n")
	}
	
	// Use pre-built template string for better performance
	builder.WriteString(contractTemplatePrompt)
	
	return builder.String()
}

// Pre-built contract template to avoid string concatenation
const contractTemplatePrompt = `

Generate the Pact contract as valid JSON following this structure:
{
  "consumer": {"name": "consumer-name"},
  "provider": {"name": "provider-name"},
  "interactions": [
    {
      "description": "interaction description",
      "request": {
        "method": "GET",
        "path": "/path",
        "headers": {},
        "body": {}
      },
      "response": {
        "status": 200,
        "headers": {},
        "body": {}
      }
    }
  ]
}
`

// parseAIResponse parses the AI response into a PactContract with optimized JSON processing
func (g *Generator) parseAIResponse(response string, _ *openapi3.T) (*PactContract, error) {
	// Extract JSON from response (AI might include additional text)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("no valid JSON found in AI response")
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var contract PactContract
	if err := json.Unmarshal([]byte(jsonStr), &contract); err != nil {
		// If direct parsing fails, try to sanitize and retry
		jsonStr = g.sanitizeJSON(jsonStr)
		if err := json.Unmarshal([]byte(jsonStr), &contract); err != nil {
			return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
		}
	}
	
	// Enhance contract with metadata using pre-calculated time
	now := time.Now()
	for i := range contract.Interactions {
		contract.Interactions[i].Metadata = InteractionMetadata{
			GeneratedAt: now,
			Source:      "AI-generated from OpenAPI spec",
		}
	}
	
	return &contract, nil
}

// sanitizeJSON attempts to fix common JSON issues in AI responses
func (g *Generator) sanitizeJSON(jsonStr string) string {
	// Remove code block markers
	jsonStr = strings.ReplaceAll(jsonStr, "```json", "")
	jsonStr = strings.ReplaceAll(jsonStr, "```", "")
	
	// Trim whitespace
	jsonStr = strings.TrimSpace(jsonStr)
	
	return jsonStr
}

// saveContract saves the contract to a file with optimized JSON marshaling
func (g *Generator) saveContract(contract *PactContract) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.config.OutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate filename using pooled builder
	builder := stringBuilderPool.Get().(*strings.Builder)
	defer stringBuilderPool.Put(builder)
	builder.Reset()
	
	builder.WriteString(strings.ToLower(contract.Consumer.Name))
	builder.WriteString("-")
	builder.WriteString(strings.ToLower(contract.Provider.Name))
	builder.WriteString("-pact.json")
	filename := builder.String()
	
	filePath := filepath.Join(g.config.OutputPath, filename)
	
	// Convert contract to JSON with pooled buffer
	jsonBuf := jsonBufferPool.Get().([]byte)
	defer jsonBufferPool.Put(jsonBuf[:0]) // Reset slice but keep capacity
	
	jsonData, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal contract to JSON: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write contract file: %w", err)
	}
	
	return filePath, nil
}

// generateTestCode generates test code for the contract
func (g *Generator) generateTestCode(ctx context.Context, contract *PactContract) (string, error) {
	if g.config.Language == "" {
		return "", nil
	}
	
	// Try to use template first
	templates := GetDefaultTemplates()
	templateKey := g.config.Language + "-" + g.config.Framework
	
	if template, exists := templates[templateKey]; exists {
		testCode, err := GenerateTestCodeFromTemplate(template, contract)
		if err == nil {
			return testCode, nil
		}
		// If template generation fails, fall back to AI generation
		slog.Warn("Template generation failed, falling back to AI", "error", err)
	}
	
	// Use AI to generate test code
	prompt := fmt.Sprintf(`Generate %s test code for the following Pact contract using %s framework.
The test should:
- Set up the Pact mock provider
- Execute the interactions
- Verify the contract
- Include proper imports and setup
- Follow best practices for %s and %s
- Include error handling and proper assertions

Pact Contract:
%s

Please provide complete, runnable test code with proper structure and comments.`, 
		g.config.Language, g.config.Framework, g.config.Language, g.config.Framework, func() string {
			json, _ := contract.ToJSON()
			return json
		}())
	
	aiResponse, err := g.ai.GenerateText(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate test code: %w", err)
	}
	
	return aiResponse.Response, nil
}

// ValidateContract validates a Pact contract using basic validation
func (g *Generator) ValidateContract(contract *PactContract) *ValidationResult {
	validator := NewEnhancedValidator(false)
	detailed := validator.ValidateDetailed(contract)
	
	// Convert detailed result to simple result for backward compatibility
	result := &ValidationResult{
		Valid:    detailed.Valid,
		Errors:   make([]string, len(detailed.Errors)),
		Warnings: make([]string, len(detailed.Warnings)),
	}
	
	for i, err := range detailed.Errors {
		result.Errors[i] = fmt.Sprintf("%s: %s", err.Location, err.Message)
	}
	
	for i, warn := range detailed.Warnings {
		result.Warnings[i] = fmt.Sprintf("%s: %s", warn.Location, warn.Message)
	}
	
	return result
}

// ValidateContractDetailed validates a Pact contract with comprehensive feedback
func (g *Generator) ValidateContractDetailed(contract *PactContract, strictMode bool) *DetailedValidationResult {
	validator := NewEnhancedValidator(strictMode)
	return validator.ValidateDetailed(contract)
}

// Close closes the generator and cleans up resources
func (g *Generator) Close() error {
	if g.ai != nil {
		g.ai.Close()
	}
	return nil
}