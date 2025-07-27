package pact

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dipjyotimetia/jarvis/pkg/engine/files"
	"github.com/dipjyotimetia/jarvis/pkg/engine/ollama"
	"github.com/getkin/kin-openapi/openapi3"
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
			fmt.Printf("Warning: failed to generate test code: %v\n", err)
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

// buildAIPrompt creates a prompt for the AI to generate Pact contracts
func (g *Generator) buildAIPrompt(_ *openapi3.T, specContent []string) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert in contract testing and Pact specification. ")
	prompt.WriteString("Generate a comprehensive Pact contract from the following OpenAPI specification.\n\n")
	
	prompt.WriteString("Requirements:\n")
	prompt.WriteString("- Create realistic interactions for each endpoint\n")
	prompt.WriteString("- Include proper HTTP methods, paths, headers, and response codes\n")
	prompt.WriteString("- Generate meaningful test data for request/response bodies\n")
	prompt.WriteString("- Include both success and error scenarios\n")
	prompt.WriteString("- Use appropriate Pact matchers for flexible matching\n")
	prompt.WriteString("- Follow Pact specification v3.0.0 format\n\n")
	
	if g.config.Language != "" {
		prompt.WriteString(fmt.Sprintf("Target language: %s\n", g.config.Language))
	}
	if g.config.Framework != "" {
		prompt.WriteString(fmt.Sprintf("Target framework: %s\n", g.config.Framework))
	}
	
	for key, value := range g.config.ExtraContext {
		prompt.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	
	prompt.WriteString("\nOpenAPI Specification:\n")
	prompt.WriteString(strings.Join(specContent, "\n"))
	
	prompt.WriteString("\n\nGenerate the Pact contract as valid JSON following this structure:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"consumer\": {\"name\": \"consumer-name\"},\n")
	prompt.WriteString("  \"provider\": {\"name\": \"provider-name\"},\n")
	prompt.WriteString("  \"interactions\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"description\": \"interaction description\",\n")
	prompt.WriteString("      \"request\": {\n")
	prompt.WriteString("        \"method\": \"GET\",\n")
	prompt.WriteString("        \"path\": \"/path\",\n")
	prompt.WriteString("        \"headers\": {},\n")
	prompt.WriteString("        \"body\": {}\n")
	prompt.WriteString("      },\n")
	prompt.WriteString("      \"response\": {\n")
	prompt.WriteString("        \"status\": 200,\n")
	prompt.WriteString("        \"headers\": {},\n")
	prompt.WriteString("        \"body\": {}\n")
	prompt.WriteString("      }\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ]\n")
	prompt.WriteString("}\n")
	
	return prompt.String()
}

// parseAIResponse parses the AI response into a PactContract
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
		// If direct parsing fails, try to fix common issues
		jsonStr = g.sanitizeJSON(jsonStr)
		if err := json.Unmarshal([]byte(jsonStr), &contract); err != nil {
			return nil, fmt.Errorf("failed to parse AI response as JSON: %w", err)
		}
	}
	
	// Enhance contract with additional metadata
	for i := range contract.Interactions {
		contract.Interactions[i].Metadata = InteractionMetadata{
			GeneratedAt: time.Now(),
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

// saveContract saves the contract to a file
func (g *Generator) saveContract(contract *PactContract) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.config.OutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate filename
	filename := fmt.Sprintf("%s-%s-pact.json", 
		strings.ToLower(contract.Consumer.Name),
		strings.ToLower(contract.Provider.Name))
	
	filePath := filepath.Join(g.config.OutputPath, filename)
	
	// Convert contract to JSON
	jsonData, err := contract.ToJSON()
	if err != nil {
		return "", fmt.Errorf("failed to convert contract to JSON: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, []byte(jsonData), 0644); err != nil {
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
		fmt.Printf("Warning: Template generation failed, falling back to AI: %v\n", err)
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