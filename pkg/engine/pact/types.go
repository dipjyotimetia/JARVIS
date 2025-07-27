package pact

import (
	"encoding/json"
	"time"
)

// PactContract represents a Pact contract file structure
type PactContract struct {
	Consumer     PactParticipant `json:"consumer"`
	Provider     PactParticipant `json:"provider"`
	Interactions []Interaction   `json:"interactions"`
	Metadata     PactMetadata    `json:"metadata"`
}

// PactParticipant represents a consumer or provider in the contract
type PactParticipant struct {
	Name string `json:"name"`
}

// Interaction represents a single interaction between consumer and provider
type Interaction struct {
	Description   string             `json:"description"`
	ProviderState string             `json:"providerState,omitempty"`
	Request       PactRequest        `json:"request"`
	Response      PactResponse       `json:"response"`
	Metadata      InteractionMetadata `json:"metadata,omitempty"`
}

// PactRequest represents the request part of an interaction
type PactRequest struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Query   map[string]interface{} `json:"query,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Body    interface{}            `json:"body,omitempty"`
}

// PactResponse represents the response part of an interaction
type PactResponse struct {
	Status  int                    `json:"status"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Body    interface{}            `json:"body,omitempty"`
}

// PactMetadata contains metadata about the contract
type PactMetadata struct {
	PactSpecification PactSpecification `json:"pactSpecification"`
	Client            ClientMetadata    `json:"client,omitempty"`
}

// PactSpecification defines the Pact specification version
type PactSpecification struct {
	Version string `json:"version"`
}

// ClientMetadata contains information about the client that generated the contract
type ClientMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InteractionMetadata contains metadata specific to an interaction
type InteractionMetadata struct {
	GeneratedAt time.Time `json:"generatedAt,omitempty"`
	Source      string    `json:"source,omitempty"`
}

// Matcher represents a Pact matcher for flexible matching
type Matcher struct {
	Match string      `json:"match"`
	Value interface{} `json:"value,omitempty"`
	Min   int         `json:"min,omitempty"`
	Max   int         `json:"max,omitempty"`
}

// GenerationConfig holds configuration for Pact generation
type GenerationConfig struct {
	ConsumerName    string            `json:"consumerName"`
	ProviderName    string            `json:"providerName"`
	OutputPath      string            `json:"outputPath"`
	SpecVersion     string            `json:"specVersion"`
	IncludeExamples bool              `json:"includeExamples"`
	Language        string            `json:"language"`
	Framework       string            `json:"framework"`
	ExtraContext    map[string]string `json:"extraContext"`
}

// ContractGenerationResult holds the result of contract generation
type ContractGenerationResult struct {
	Contract     *PactContract `json:"contract"`
	FilePath     string        `json:"filePath"`
	TestCode     string        `json:"testCode,omitempty"`
	Language     string        `json:"language"`
	Framework    string        `json:"framework"`
	GeneratedAt  time.Time     `json:"generatedAt"`
	SourceSpec   string        `json:"sourceSpec"`
	Interactions int           `json:"interactionCount"`
}

// ValidationResult holds the result of contract validation
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ToJSON converts the PactContract to JSON string
func (p *PactContract) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AddInteraction adds a new interaction to the contract
func (p *PactContract) AddInteraction(interaction Interaction) {
	p.Interactions = append(p.Interactions, interaction)
}

// SetMetadata sets the contract metadata
func (p *PactContract) SetMetadata(version, clientName, clientVersion string) {
	p.Metadata = PactMetadata{
		PactSpecification: PactSpecification{Version: version},
		Client: ClientMetadata{
			Name:    clientName,
			Version: clientVersion,
		},
	}
}