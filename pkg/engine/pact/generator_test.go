package pact

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPactContractGeneration(t *testing.T) {
	// Skip if we don't have Ollama running
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	config := &GenerationConfig{
		ConsumerName:    "test-consumer",
		ProviderName:    "test-provider",
		OutputPath:      "./test-output",
		SpecVersion:     "3.0.0",
		IncludeExamples: false,
		Language:        "javascript",
		Framework:       "jest",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	generator, err := NewGenerator(ctx, config)
	if err != nil {
		t.Skipf("Cannot create generator (Ollama may not be available): %v", err)
	}
	defer generator.Close()

	// Test basic contract creation
	contract := &PactContract{
		Consumer: PactParticipant{Name: "test-consumer"},
		Provider: PactParticipant{Name: "test-provider"},
		Interactions: []Interaction{
			{
				Description: "Test interaction",
				Request: PactRequest{
					Method: "GET",
					Path:   "/test",
				},
				Response: PactResponse{
					Status: 200,
					Body:   map[string]interface{}{"message": "success"},
				},
			},
		},
	}
	contract.SetMetadata("3.0.0", "jarvis-test", "1.0.0")

	// Test validation
	validation := generator.ValidateContract(contract)
	if !validation.Valid {
		t.Errorf("Contract validation failed: %v", validation.Errors)
	}

	// Test detailed validation
	detailedValidation := generator.ValidateContractDetailed(contract, false)
	if !detailedValidation.Valid {
		t.Errorf("Detailed contract validation failed: %v", detailedValidation.Errors)
	}

	// Test JSON generation
	jsonStr, err := contract.ToJSON()
	if err != nil {
		t.Errorf("Failed to generate JSON: %v", err)
	}

	if jsonStr == "" {
		t.Error("Generated JSON is empty")
	}

	// Clean up test output directory
	os.RemoveAll("./test-output")
}

func TestPactValidation(t *testing.T) {
	validator := NewEnhancedValidator(false)

	// Test invalid contract
	invalidContract := &PactContract{
		Consumer: PactParticipant{Name: ""},
		Provider: PactParticipant{Name: "test-provider"},
		Interactions: []Interaction{
			{
				Description: "",
				Request: PactRequest{
					Method: "",
					Path:   "",
				},
				Response: PactResponse{
					Status: 0,
				},
			},
		},
	}

	result := validator.ValidateDetailed(invalidContract)
	if result.Valid {
		t.Error("Expected validation to fail for invalid contract")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected validation errors")
	}

	// Test valid contract
	validContract := &PactContract{
		Consumer: PactParticipant{Name: "consumer"},
		Provider: PactParticipant{Name: "provider"},
		Interactions: []Interaction{
			{
				Description: "Valid interaction",
				Request: PactRequest{
					Method: "GET",
					Path:   "/api/test",
				},
				Response: PactResponse{
					Status: 200,
					Body:   map[string]interface{}{"result": "success"},
				},
			},
		},
	}
	validContract.SetMetadata("3.0.0", "test", "1.0.0")

	validResult := validator.ValidateDetailed(validContract)
	if !validResult.Valid {
		t.Errorf("Expected validation to pass for valid contract: %v", validResult.Errors)
	}
}

func TestTemplateGeneration(t *testing.T) {
	templates := GetDefaultTemplates()

	if len(templates) == 0 {
		t.Error("Expected default templates to be available")
	}

	// Test JavaScript Jest template
	jsTemplate, exists := templates["javascript-jest"]
	if !exists {
		t.Error("Expected JavaScript Jest template to exist")
	}

	contract := &PactContract{
		Consumer: PactParticipant{Name: "web-app"},
		Provider: PactParticipant{Name: "api-service"},
		Interactions: []Interaction{
			{
				Description: "Get user data",
				Request: PactRequest{
					Method: "GET",
					Path:   "/users/123",
				},
				Response: PactResponse{
					Status: 200,
					Body:   map[string]interface{}{"id": 123, "name": "John Doe"},
				},
			},
		},
	}

	testCode, err := GenerateTestCodeFromTemplate(jsTemplate, contract)
	if err != nil {
		t.Errorf("Failed to generate test code from template: %v", err)
	}

	if testCode == "" {
		t.Error("Generated test code is empty")
	}

	// Check that placeholders were replaced
	if strings.Contains(testCode, "{{CONSUMER_NAME}}") {
		t.Error("Template placeholders were not replaced")
	}
}