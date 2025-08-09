package pact

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// DetailedValidationResult provides comprehensive validation feedback
type DetailedValidationResult struct {
	Valid              bool                    `json:"valid"`
	Errors             []ValidationError       `json:"errors,omitempty"`
	Warnings           []ValidationWarning     `json:"warnings,omitempty"`
	Suggestions        []ValidationSuggestion  `json:"suggestions,omitempty"`
	InteractionResults []InteractionValidation `json:"interactionResults,omitempty"`
	Metadata           ValidationMetadata      `json:"metadata"`
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Location   string `json:"location"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Location   string `json:"location"`
	Suggestion string `json:"suggestion,omitempty"`
}

// ValidationSuggestion represents an improvement suggestion
type ValidationSuggestion struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Location string `json:"location"`
	Example  string `json:"example,omitempty"`
}

// InteractionValidation represents validation results for a single interaction
type InteractionValidation struct {
	Index       int                 `json:"index"`
	Description string              `json:"description"`
	Valid       bool                `json:"valid"`
	Errors      []ValidationError   `json:"errors,omitempty"`
	Warnings    []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationMetadata contains metadata about the validation process
type ValidationMetadata struct {
	TotalInteractions   int    `json:"totalInteractions"`
	ValidInteractions   int    `json:"validInteractions"`
	InvalidInteractions int    `json:"invalidInteractions"`
	SpecVersion         string `json:"specVersion"`
	ValidationTime      string `json:"validationTime"`
}

// EnhancedValidator provides comprehensive Pact contract validation
type EnhancedValidator struct {
	strictMode bool
	rules      []ValidationRule
}

// ValidationRule defines a validation rule
type ValidationRule struct {
	Name        string
	Description string
	Validator   func(*PactContract) []ValidationError
}

// NewEnhancedValidator creates a new enhanced validator
func NewEnhancedValidator(strictMode bool) *EnhancedValidator {
	validator := &EnhancedValidator{
		strictMode: strictMode,
		rules:      getDefaultValidationRules(),
	}

	if strictMode {
		validator.rules = append(validator.rules, getStrictValidationRules()...)
	}

	return validator
}

// ValidateDetailed performs comprehensive validation of a Pact contract
func (v *EnhancedValidator) ValidateDetailed(contract *PactContract) *DetailedValidationResult {
	result := &DetailedValidationResult{
		Valid:              true,
		Errors:             []ValidationError{},
		Warnings:           []ValidationWarning{},
		Suggestions:        []ValidationSuggestion{},
		InteractionResults: []InteractionValidation{},
		Metadata: ValidationMetadata{
			TotalInteractions: len(contract.Interactions),
			SpecVersion:       contract.Metadata.PactSpecification.Version,
		},
	}

	// Run all validation rules
	for _, rule := range v.rules {
		errors := rule.Validator(contract)
		result.Errors = append(result.Errors, errors...)
		if len(errors) > 0 {
			result.Valid = false
		}
	}

	// Validate each interaction
	validInteractions := 0
	for i, interaction := range contract.Interactions {
		interactionResult := v.validateInteraction(i, &interaction)
		result.InteractionResults = append(result.InteractionResults, interactionResult)

		if interactionResult.Valid {
			validInteractions++
		} else {
			result.Valid = false
		}
	}

	result.Metadata.ValidInteractions = validInteractions
	result.Metadata.InvalidInteractions = len(contract.Interactions) - validInteractions

	// Generate suggestions
	result.Suggestions = v.generateSuggestions(contract, result)

	return result
}

// validateInteraction validates a single interaction
func (v *EnhancedValidator) validateInteraction(index int, interaction *Interaction) InteractionValidation {
	result := InteractionValidation{
		Index:       index,
		Description: interaction.Description,
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
	}

	location := fmt.Sprintf("interactions[%d]", index)

	// Validate description
	if strings.TrimSpace(interaction.Description) == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:       "MISSING_DESCRIPTION",
			Message:    "Interaction description is required",
			Location:   location + ".description",
			Severity:   "error",
			Suggestion: "Add a clear, descriptive name for this interaction",
		})
		result.Valid = false
	}

	// Validate request
	if err := v.validateRequest(interaction.Request, location+".request"); err != nil {
		result.Errors = append(result.Errors, *err)
		result.Valid = false
	}

	// Validate response (classify severity)
	if err := v.validateResponse(interaction.Response, location+".response"); err != nil {
		if strings.EqualFold(err.Severity, "warning") {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:       err.Code,
				Message:    err.Message,
				Location:   err.Location,
				Suggestion: err.Suggestion,
			})
		} else {
			result.Errors = append(result.Errors, *err)
			result.Valid = false
		}
	}

	// Check for best practices
	warnings := v.checkInteractionBestPractices(interaction, location)
	result.Warnings = append(result.Warnings, warnings...)

	return result
}

// validateRequest validates the request part of an interaction
func (v *EnhancedValidator) validateRequest(request PactRequest, location string) *ValidationError {
	// Validate HTTP method
	if request.Method == "" {
		return &ValidationError{
			Code:       "MISSING_METHOD",
			Message:    "Request method is required",
			Location:   location + ".method",
			Severity:   "error",
			Suggestion: "Specify a valid HTTP method (GET, POST, PUT, DELETE, etc.)",
		}
	}

	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	if !contains(validMethods, strings.ToUpper(request.Method)) {
		return &ValidationError{
			Code:       "INVALID_METHOD",
			Message:    fmt.Sprintf("Invalid HTTP method: %s", request.Method),
			Location:   location + ".method",
			Severity:   "error",
			Suggestion: fmt.Sprintf("Use one of: %s", strings.Join(validMethods, ", ")),
		}
	}

	// Validate path
	if request.Path == "" {
		return &ValidationError{
			Code:       "MISSING_PATH",
			Message:    "Request path is required",
			Location:   location + ".path",
			Severity:   "error",
			Suggestion: "Specify a valid URL path starting with '/'",
		}
	}

	if !strings.HasPrefix(request.Path, "/") {
		return &ValidationError{
			Code:       "INVALID_PATH",
			Message:    "Request path must start with '/'",
			Location:   location + ".path",
			Severity:   "error",
			Suggestion: "Ensure path starts with '/' character",
		}
	}

	// Validate URL encoding
	if _, err := url.Parse(request.Path); err != nil {
		return &ValidationError{
			Code:       "INVALID_URL",
			Message:    fmt.Sprintf("Invalid URL path: %v", err),
			Location:   location + ".path",
			Severity:   "error",
			Suggestion: "Ensure path is properly URL-encoded",
		}
	}

	return nil
}

// validateResponse validates the response part of an interaction
func (v *EnhancedValidator) validateResponse(response PactResponse, location string) *ValidationError {
	// Validate status code
	if response.Status == 0 {
		return &ValidationError{
			Code:       "MISSING_STATUS",
			Message:    "Response status code is required",
			Location:   location + ".status",
			Severity:   "error",
			Suggestion: "Specify a valid HTTP status code (200, 404, 500, etc.)",
		}
	}

	if response.Status < 100 || response.Status > 599 {
		return &ValidationError{
			Code:       "INVALID_STATUS",
			Message:    fmt.Sprintf("Invalid HTTP status code: %d", response.Status),
			Location:   location + ".status",
			Severity:   "error",
			Suggestion: "Use a valid HTTP status code between 100-599",
		}
	}

	// Validate that response has content-type if body is present
	if response.Body != nil {
		if headers, ok := response.Headers["Content-Type"]; !ok || headers == nil {
			return &ValidationError{
				Code:       "MISSING_CONTENT_TYPE",
				Message:    "Content-Type header is recommended when response has a body",
				Location:   location + ".headers",
				Severity:   "warning",
				Suggestion: "Add Content-Type header to specify response format",
			}
		}
	}

	return nil
}

// checkInteractionBestPractices checks for best practice violations
func (v *EnhancedValidator) checkInteractionBestPractices(interaction *Interaction, location string) []ValidationWarning {
	var warnings []ValidationWarning

	// Check description quality
	if len(interaction.Description) < 10 {
		warnings = append(warnings, ValidationWarning{
			Code:       "SHORT_DESCRIPTION",
			Message:    "Interaction description is very short",
			Location:   location + ".description",
			Suggestion: "Provide a more descriptive name that explains what this interaction tests",
		})
	}

	// Check for provider state usage
	if strings.TrimSpace(interaction.ProviderState) == "" && v.strictMode {
		warnings = append(warnings, ValidationWarning{
			Code:       "MISSING_PROVIDER_STATE",
			Message:    "Provider state is not specified",
			Location:   location + ".providerState",
			Suggestion: "Consider adding provider state to make tests more reliable",
		})
	}

	// Check request headers
	if interaction.Request.Headers == nil || len(interaction.Request.Headers) == 0 {
		warnings = append(warnings, ValidationWarning{
			Code:       "NO_REQUEST_HEADERS",
			Message:    "No request headers specified",
			Location:   location + ".request.headers",
			Suggestion: "Consider adding relevant headers like Accept, Authorization, etc.",
		})
	}

	return warnings
}

// generateSuggestions generates improvement suggestions based on validation results
func (v *EnhancedValidator) generateSuggestions(contract *PactContract, _ *DetailedValidationResult) []ValidationSuggestion {
	var suggestions []ValidationSuggestion

	// Suggest adding more interactions if only one exists
	if len(contract.Interactions) == 1 {
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "completeness",
			Message:  "Consider adding more interactions to improve test coverage",
			Location: "interactions",
			Example:  "Add interactions for error scenarios, edge cases, and different endpoints",
		})
	}

	// Suggest adding error scenarios if all responses are successful
	allSuccess := true
	for _, interaction := range contract.Interactions {
		if interaction.Response.Status >= 400 {
			allSuccess = false
			break
		}
	}

	if allSuccess && len(contract.Interactions) > 1 {
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "coverage",
			Message:  "Consider adding error scenario interactions",
			Location: "interactions",
			Example:  "Add interactions for 400, 404, 500 status codes to test error handling",
		})
	}

	// Suggest using matchers for flexible matching
	suggestions = append(suggestions, ValidationSuggestion{
		Type:     "improvement",
		Message:  "Consider using Pact matchers for more flexible contract matching",
		Location: "interactions[*].response.body",
		Example:  "Use type matchers instead of exact values: {\"id\": {\"match\": \"type\"}}",
	})

	return suggestions
}

// getDefaultValidationRules returns the default set of validation rules
func getDefaultValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Name:        "Consumer Name Required",
			Description: "Validates that consumer name is specified",
			Validator: func(contract *PactContract) []ValidationError {
				if strings.TrimSpace(contract.Consumer.Name) == "" {
					return []ValidationError{{
						Code:       "MISSING_CONSUMER",
						Message:    "Consumer name is required",
						Location:   "consumer.name",
						Severity:   "error",
						Suggestion: "Specify the name of the service consuming the API",
					}}
				}
				return nil
			},
		},
		{
			Name:        "Provider Name Required",
			Description: "Validates that provider name is specified",
			Validator: func(contract *PactContract) []ValidationError {
				if strings.TrimSpace(contract.Provider.Name) == "" {
					return []ValidationError{{
						Code:       "MISSING_PROVIDER",
						Message:    "Provider name is required",
						Location:   "provider.name",
						Severity:   "error",
						Suggestion: "Specify the name of the service providing the API",
					}}
				}
				return nil
			},
		},
		{
			Name:        "Interactions Required",
			Description: "Validates that at least one interaction exists",
			Validator: func(contract *PactContract) []ValidationError {
				if len(contract.Interactions) == 0 {
					return []ValidationError{{
						Code:       "NO_INTERACTIONS",
						Message:    "At least one interaction is required",
						Location:   "interactions",
						Severity:   "error",
						Suggestion: "Add one or more interactions that define the expected API behavior",
					}}
				}
				return nil
			},
		},
	}
}

// getStrictValidationRules returns additional rules for strict mode
func getStrictValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Name:        "Semantic Versioning",
			Description: "Validates that spec version follows semantic versioning",
			Validator: func(contract *PactContract) []ValidationError {
				version := contract.Metadata.PactSpecification.Version
				if !isValidSemVer(version) {
					return []ValidationError{{
						Code:       "INVALID_VERSION",
						Message:    fmt.Sprintf("Pact specification version should follow semantic versioning: %s", version),
						Location:   "metadata.pactSpecification.version",
						Severity:   "warning",
						Suggestion: "Use a valid semantic version like '3.0.0'",
					}}
				}
				return nil
			},
		},
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidSemVer(version string) bool {
	// Simple regex for semantic versioning
	regex := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	return regex.MatchString(version)
}

// ValidateJSON validates that a string contains valid JSON
func ValidateJSON(jsonStr string) error {
	var temp interface{}
	return json.Unmarshal([]byte(jsonStr), &temp)
}
